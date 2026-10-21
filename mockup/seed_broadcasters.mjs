import fs from 'fs';
import path from 'path';

const API_BASE = 'http://localhost:4000';
const CSV_FILE = 'mockup/profiles.csv';
const AVATARS_DIR = 'mockup/ladies_profiles_avatars';

function parseCSV(text) {
  const lines = [];
  let row = [];
  let inQuotes = false;
  let current = '';

  for (let i = 0; i < text.length; i++) {
    const char = text[i];
    const next = text[i + 1];

    if (char === '"') {
      if (inQuotes && next === '"') {
        current += '"';
        i++;
      } else {
        inQuotes = !inQuotes;
      }
    } else if (char === ',' && !inQuotes) {
      row.push(current);
      current = '';
    } else if ((char === '\r' || char === '\n') && !inQuotes) {
      if (char === '\r' && next === '\n') i++;
      row.push(current);
      current = '';
      if (row.length > 1 || (row.length === 1 && row[0] !== '')) {
        lines.push(row);
      }
      row = [];
    } else {
      current += char;
    }
  }
  if (current || row.length > 0) {
    row.push(current);
    lines.push(row);
  }
  return lines;
}

async function uploadImage(filePath, description) {
  const fileBytes = fs.readFileSync(filePath);
  const base64Data = fileBytes.toString('base64');

  const res = await fetch(`${API_BASE}/image`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      description,
      data: base64Data,
      content_type: 'image/png',
    }),
  });

  if (!res.ok) {
    const errText = await res.text();
    throw new Error(`Upload failed for ${filePath}: ${res.status} ${errText}`);
  }

  const data = await res.json();
  return data.url;
}

async function run() {
  console.log('--- Inizio seeding/aggiornamento profili broadcaster ---');
  const csvContent = fs.readFileSync(CSV_FILE, 'utf-8');
  const rows = parseCSV(csvContent);
  const dataRows = rows.slice(1); // skip header

  console.log(`Trovati ${dataRows.length} profili nel file CSV.`);

  for (let index = 0; index < dataRows.length; index++) {
    const row = dataRows[index];
    const id = row[0];
    const fullName = row[1];
    const age = parseInt(row[2], 10);
    const city = row[3];
    const state = row[4];
    const bio = row[5];
    const folder = row[6];

    const slug = fullName.toLowerCase().replace(/[^a-z0-9]+/g, '.').replace(/^\.+|\.+$/g, '');
    const email = `${slug}@broadcaster.local`;
    const password = 'Password123!';

    console.log(`\n[${index + 1}/${dataRows.length}] Elaborazione: ${fullName} (${folder}) - Email: ${email}`);

    // Controlla se il broadcaster esiste già
    const searchRes = await fetch(`${API_BASE}/broadcasters?search=${encodeURIComponent(fullName)}`);
    const searchData = searchRes.ok ? await searchRes.json() : { broadcasters: [] };
    const existingBc = searchData.broadcasters?.find((b) => b.display_name === fullName);

    let avatarUrl = '';
    const photos = [];

    if (!existingBc || !existingBc.photos || existingBc.photos.length === 0) {
      // 1. Upload photo_1.png per l'avatar
      const photo1Path = path.join(AVATARS_DIR, folder, 'photo_1.png');
      if (fs.existsSync(photo1Path)) {
        console.log(`  Caricamento avatar (${photo1Path})...`);
        avatarUrl = await uploadImage(photo1Path, `${fullName} - Avatar`);
      }

      // 2. Upload photo_2.png .. photo_7.png per le foto del broadcaster
      for (let p = 2; p <= 7; p++) {
        const photoPath = path.join(AVATARS_DIR, folder, `photo_${p}.png`);
        if (fs.existsSync(photoPath)) {
          console.log(`  Caricamento photo_${p}...`);
          const photoUrl = await uploadImage(photoPath, `${fullName} - Photo ${p}`);
          photos.push({
            image_url: photoUrl,
            alt: `${fullName} - Photo ${p}`,
          });
        }
      }
    } else {
      console.log(`  Broadcaster già presente con ${existingBc.photos.length} foto. Uso foto esistenti.`);
      for (const ph of existingBc.photos) {
        photos.push(ph);
      }
    }

    // 3. Registrazione se nuovo
    let userId = existingBc?.user_id || null;
    if (!existingBc) {
      const regRes = await fetch(`${API_BASE}/profiles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email,
          display_name: fullName,
          avatar_url: avatarUrl,
          password,
        }),
      });
      if (regRes.ok) {
        const userProfile = await regRes.json();
        userId = userProfile.user_id;
        console.log(`  Utente creato: ${userId}`);
      }
    }

    // 4. Login
    const loginRes = await fetch(`${API_BASE}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        user_email: email,
        password,
      }),
    });

    if (!loginRes.ok) {
      console.error(`  Errore login per ${email}`);
      continue;
    }

    const { token } = await loginRes.json();

    const payload = {
      display_name: fullName,
      category: 'Beauty',
      category_color: '#6366f1',
      bio,
      certified: true,
      availability: 'available',
      phone: '',
      services: ['Live Video', 'Chat', 'Private Room'],
      languages: ['English'],
      attributes: [
        { label: 'Age', value: String(age) },
        { label: 'City', value: city },
        { label: 'State', value: state },
      ],
      photos,
      enabled: true,
      used: true,
    };

    if (existingBc) {
      // Aggiorna
      const putRes = await fetch(`${API_BASE}/broadcasters/id/${existingBc.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });
      console.log(`  Aggiornamento broadcaster ${existingBc.id}: status ${putRes.status}`);
    } else {
      // Crea
      const bcRes = await fetch(`${API_BASE}/broadcasters`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });
      if (bcRes.ok) {
        const bcData = await bcRes.json();
        console.log(`  Broadcaster creato con successo! ID: ${bcData.id}`);
      }
    }
  }

  console.log('\n--- Completato con successo! ---');
}

run().catch((err) => {
  console.error('Errore:', err);
  process.exit(1);
});
