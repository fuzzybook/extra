import fs from 'node:fs';
import path from 'node:path';
import ts from 'typescript';

const rootDir = process.cwd();
const clientPath = path.join(rootDir, 'src/encore/client.ts');
const outputPath = path.join(rootDir, 'src/encore/zod.ts');

const clientSource = fs.readFileSync(clientPath, 'utf8');
const clientFile = ts.createSourceFile(clientPath, clientSource, ts.ScriptTarget.Latest, true);
const existingSource = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, 'utf8') : '';
const existingFile = ts.createSourceFile(outputPath, existingSource, ts.ScriptTarget.Latest, true);

const schemas = collectEncoreSchemas(clientFile);
const existingSchemas = collectExistingSchemas(existingFile, existingSource);
const generated = renderOutput(schemas, existingSchemas);

fs.writeFileSync(outputPath, generated);
console.log(`Synced ${schemas.length} Zod schemas to ${path.relative(rootDir, outputPath)}`);

function collectEncoreSchemas(sourceFile) {
  const found = [];
  const typeAliases = new Map();
  const writableParamTypes = collectWritableParamTypes(sourceFile);

  for (const statement of sourceFile.statements) {
    if (!isExportedNamespace(statement)) continue;
    const namespaceName = statement.name.text;
    const body = statement.body;
    if (!body || !ts.isModuleBlock(body)) continue;

    for (const child of body.statements) {
      if (ts.isTypeAliasDeclaration(child) && isExported(child)) {
        typeAliases.set(`${namespaceName}.${child.name.text}`, child.type);
      }
    }
  }

  for (const statement of sourceFile.statements) {
    if (!isExportedNamespace(statement)) continue;
    const namespaceName = statement.name.text;
    const body = statement.body;
    if (!body || !ts.isModuleBlock(body)) continue;

    for (const child of body.statements) {
      const qualifiedName = `${namespaceName}.${child.name?.text ?? ''}`;
      if (!writableParamTypes.has(qualifiedName)) continue;

      if (ts.isInterfaceDeclaration(child) && isExported(child)) {
        found.push({
          kind: 'object',
          namespaceName,
          typeName: child.name.text,
          schemaName: schemaName(namespaceName, child.name.text),
          fields: child.members
            .filter(ts.isPropertySignature)
            .map((member) => ({
              name: propertyName(member.name),
              zod: zodForType(member.type, namespaceName, typeAliases),
            }))
            .filter((field) => field.name),
        });
      }

      if (ts.isTypeAliasDeclaration(child) && isExported(child)) {
        found.push({
          kind: 'alias',
          namespaceName,
          typeName: child.name.text,
          schemaName: schemaName(namespaceName, child.name.text),
          zod: zodForType(child.type, namespaceName, typeAliases),
        });
      }
    }
  }

  return found;
}

function collectWritableParamTypes(sourceFile) {
  const found = new Set();

  for (const statement of sourceFile.statements) {
    if (!isExportedNamespace(statement)) continue;
    const namespaceName = statement.name.text;
    const body = statement.body;
    if (!body || !ts.isModuleBlock(body)) continue;

    for (const child of body.statements) {
      if (!ts.isClassDeclaration(child) || child.name?.text !== 'ServiceClient') continue;

      for (const member of child.members) {
        if (!ts.isMethodDeclaration(member)) continue;
        const httpMethod = writableHttpMethod(member);
        if (!httpMethod) continue;

        const bodyParamName = jsonStringifiedParamName(member);
        if (!bodyParamName) continue;

        const param = member.parameters.find((candidate) => propertyName(candidate.name) === bodyParamName);
        if (!param?.type || !ts.isTypeReferenceNode(param.type)) continue;

        const ref = typeNameText(param.type.typeName);
        const qualifiedRef = ref.includes('.') ? ref : `${namespaceName}.${ref}`;
        found.add(qualifiedRef);
      }
    }
  }

  return found;
}

function writableHttpMethod(method) {
  let result = '';

  visit(method.body);
  return result;

  function visit(node) {
    if (result || !node) return;

    if (ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression)) {
      const callName = node.expression.name.text;
      if (callName === 'callTypedAPI' || callName === 'callAPI') {
        const [methodArg] = node.arguments;
        if (methodArg && ts.isStringLiteral(methodArg) && ['POST', 'PUT'].includes(methodArg.text)) {
          result = methodArg.text;
          return;
        }
      }
    }

    ts.forEachChild(node, visit);
  }
}

function jsonStringifiedParamName(method) {
  let result = '';

  visit(method.body);
  return result;

  function visit(node) {
    if (result || !node) return;

    if (
      ts.isCallExpression(node) &&
      ts.isPropertyAccessExpression(node.expression) &&
      ts.isIdentifier(node.expression.expression) &&
      node.expression.expression.text === 'JSON' &&
      node.expression.name.text === 'stringify'
    ) {
      const [arg] = node.arguments;
      if (arg && ts.isIdentifier(arg)) {
        result = arg.text;
        return;
      }
    }

    ts.forEachChild(node, visit);
  }
}

function collectExistingSchemas(sourceFile, sourceText) {
  const result = new Map();

  for (const statement of sourceFile.statements) {
    if (!ts.isVariableStatement(statement) || !isExported(statement)) continue;

    for (const declaration of statement.declarationList.declarations) {
      if (!ts.isIdentifier(declaration.name) || !declaration.initializer) continue;
      if (!declaration.name.text.endsWith('Schema')) continue;

      const objectLiteral = zodObjectLiteral(declaration.initializer);
      if (!objectLiteral) continue;

      const fields = new Map();
      for (const prop of objectLiteral.properties) {
        if (!ts.isPropertyAssignment(prop)) continue;
        const name = propertyName(prop.name);
        if (!name) continue;
        fields.set(name, prop.initializer.getText(sourceFile));
      }

      result.set(declaration.name.text, { fields });
    }
  }

  if (!sourceText.trim()) return result;
  return result;
}

function renderOutput(schemas, existingSchemas) {
  const lines = [
    '// Code synced from src/encore/client.ts by frontend/tools/zod-sync.mjs.',
    '// Existing field validators are preserved when this file is synced again.',
    "import { z } from 'zod';",
    '',
  ];

  for (const schema of schemas) {
    if (schema.kind === 'alias') {
      lines.push(`export const ${schema.schemaName} = ${schema.zod};`);
      lines.push('');
      continue;
    }

    const existing = existingSchemas.get(schema.schemaName);
    const currentNames = new Set(schema.fields.map((field) => field.name));
    lines.push(`export const ${schema.schemaName} = z.object({`);
    for (const field of schema.fields) {
      const expression = existing?.fields.get(field.name) ?? field.zod;
      lines.push(`  ${quoteKey(field.name)}: ${expression},`);
    }
    for (const [name, expression] of existing?.fields ?? []) {
      if (currentNames.has(name)) continue;
      lines.push(`  // TODO: no longer present in Encore type ${schema.namespaceName}.${schema.typeName}`);
      lines.push(`  ${quoteKey(name)}: ${expression},`);
    }
    lines.push('});');
    lines.push('');
  }

  return `${lines.join('\n').trimEnd()}\n`;
}

function zodObjectLiteral(initializer) {
  if (!ts.isCallExpression(initializer)) return null;
  if (!ts.isPropertyAccessExpression(initializer.expression)) return null;
  if (initializer.expression.name.text !== 'object') return null;
  if (!ts.isIdentifier(initializer.expression.expression)) return null;
  if (initializer.expression.expression.text !== 'z') return null;
  const [arg] = initializer.arguments;
  return arg && ts.isObjectLiteralExpression(arg) ? arg : null;
}

function zodForType(type, namespaceName, typeAliases) {
  if (!type) return 'z.unknown()';

  if (type.kind === ts.SyntaxKind.StringKeyword) return 'z.string()';
  if (type.kind === ts.SyntaxKind.NumberKeyword) return 'z.number()';
  if (type.kind === ts.SyntaxKind.BooleanKeyword) return 'z.boolean()';
  if (type.kind === ts.SyntaxKind.AnyKeyword) return 'z.any()';
  if (type.kind === ts.SyntaxKind.UnknownKeyword) return 'z.unknown()';

  if (ts.isArrayTypeNode(type)) {
    return `z.array(${zodForType(type.elementType, namespaceName, typeAliases)})`;
  }

  if (ts.isUnionTypeNode(type)) {
    const literals = type.types.filter(ts.isLiteralTypeNode);
    if (literals.length === type.types.length && literals.length > 0) {
      return `z.union([${literals.map((literal) => zodLiteral(literal)).join(', ')}])`;
    }
    return 'z.unknown()';
  }

  if (ts.isTypeLiteralNode(type)) {
    const fields = type.members
      .filter(ts.isPropertySignature)
      .map((member) => `${quoteKey(propertyName(member.name))}: ${zodForType(member.type, namespaceName, typeAliases)}`);
    return `z.object({ ${fields.join(', ')} })`;
  }

  if (ts.isTypeReferenceNode(type)) {
    const ref = typeNameText(type.typeName);
    const qualifiedRef = ref.includes('.') ? ref : `${namespaceName}.${ref}`;
    const aliasType = typeAliases.get(qualifiedRef);
    if (aliasType) return zodForType(aliasType, namespaceName, typeAliases);
    return `z.lazy(() => ${schemaNameFromReference(ref, namespaceName)})`;
  }

  return 'z.unknown()';
}

function zodLiteral(literal) {
  const node = literal.literal;
  if (ts.isStringLiteral(node)) return `z.literal(${JSON.stringify(node.text)})`;
  if (ts.isNumericLiteral(node)) return `z.literal(${node.text})`;
  if (node.kind === ts.SyntaxKind.TrueKeyword) return 'z.literal(true)';
  if (node.kind === ts.SyntaxKind.FalseKeyword) return 'z.literal(false)';
  return 'z.unknown()';
}

function schemaNameFromReference(ref, namespaceName) {
  if (ref.includes('.')) {
    const [ns, name] = ref.split('.');
    return schemaName(ns, name);
  }
  return schemaName(namespaceName, ref);
}

function schemaName(namespaceName, typeName) {
  return `${pascal(namespaceName)}${pascal(typeName)}Schema`;
}

function propertyName(name) {
  if (!name) return '';
  if (ts.isIdentifier(name) || ts.isStringLiteral(name) || ts.isNumericLiteral(name)) return name.text;
  return '';
}

function quoteKey(key) {
  return /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(key) ? key : JSON.stringify(key);
}

function typeNameText(name) {
  if (ts.isIdentifier(name)) return name.text;
  if (ts.isQualifiedName(name)) return `${typeNameText(name.left)}.${name.right.text}`;
  return 'unknown';
}

function pascal(value) {
  return value
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((part) => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join('');
}

function isExportedNamespace(node) {
  return ts.isModuleDeclaration(node) && isExported(node) && ts.isIdentifier(node.name);
}

function isExported(node) {
  return Boolean(node.modifiers?.some((modifier) => modifier.kind === ts.SyntaxKind.ExportKeyword));
}
