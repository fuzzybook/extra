export type SupportedLocale = 'it' | 'fr' | 'de' | 'en';

export interface MessageSchema {
  common: {
    backHome: string;
    dashboard: string;
    online: string;
    loading: string;
  };
  menu: {
    home: string;
    one: string;
    two: string;
    three: string;
    webcam: string;
    contact: string;
  };
  actions: {
    advertise: string;
    login: string;
    goToDashboard: string;
  };
  home: {
    title: string;
    subtitle: string;
  };
  one: {
    title: string;
    subtitle: string;
  };
  two: {
    title: string;
    subtitle: string;
  };
  three: {
    title: string;
    subtitle: string;
  };
  webcam: {
    title: string;
    subtitle: string;
    status: string;
  };
  contact: {
    title: string;
    subtitle: string;
    supportText: string;
  };
  advertise: {
    title: string;
    subtitle: string;
    feature1: string;
    feature2: string;
    startNow: string;
  };
  login: {
    title: string;
    subtitle: string;
    emailLabel: string;
    emailPlaceholder: string;
    passwordLabel: string;
    button: string;
    forgotPassword: string;
    invalidCredentials: string;
    genericError: string;
    resetTitle: string;
    resetHint: string;
    resetButton: string;
    resetSent: string;
    backToLogin: string;
    registerTitle: string;
    registerButton: string;
    confirmPasswordLabel: string;
    passwordMismatch: string;
    emailInvalid: string;
    emailUnavailable: string;
    emailDomainInvalid: string;
    registerSuccess: string;
    languageLabel: string;
    statusLabel: string;
    nicknameLabel: string;
    nameLabel: string;
    phoneLabel: string;
    countryLabel: string;
    addressLabel: string;
    optional: string;
    newsletter: string;
    acceptTerms: string;
    termsLink: string;
    termsRequired: string;
    statusRequired: string;
    phoneInvalid: string;
    broadcasterHint: string;
    memberTypes: {
      visitor: string;
      broadcaster: string;
      affiliate: string;
      advertiser: string;
    };
  };
  terms: {
    title: string;
    subtitle: string;
    body: string;
    updated: string;
  };
}

const it: MessageSchema = {
  common: {
    backHome: 'Torna alla Home',
    dashboard: 'Dashboard',
    online: 'Online',
    loading: 'Caricamento…',
  },
  menu: {
    home: 'Home',
    one: 'ONE',
    two: 'TWO',
    three: 'THREE',
    webcam: 'Webcam',
    contact: 'Contatti',
  },
  actions: {
    advertise: 'Advertise',
    login: 'Login',
    goToDashboard: 'Vai alla Dashboard',
  },
  home: {
    title: 'PublicSite',
    subtitle: 'Pagina pubblica statica (SSG) in modalità ibrida.',
  },
  one: {
    title: 'ONE',
    subtitle: 'Sezione esclusiva ONE del portale pubblico.',
  },
  two: {
    title: 'TWO',
    subtitle: 'Sezione TWO del portale pubblico.',
  },
  three: {
    title: 'THREE',
    subtitle: 'Sezione THREE del portale pubblico.',
  },
  webcam: {
    title: 'Webcam',
    subtitle: 'Sezione Live Streaming e Webcam interattive.',
    status: 'Canale Streaming Pronto',
  },
  contact: {
    title: 'Contatti',
    subtitle: 'Mettiti in contatto con il nostro team di supporto.',
    supportText: 'Assistenza attiva 24/7',
  },
  advertise: {
    title: 'Pubblicità',
    subtitle: 'Promuovi la tua attività e aumenta la visibilità sul nostro network.',
    feature1: 'Banner in posizioni ad alto rendimento (Top & Sidebar)',
    feature2: 'Statistiche dettagliate di click e visualizzazioni',
    startNow: 'Inizia ora',
  },
  login: {
    title: 'Login',
    subtitle: 'Accedi al tuo account o creane uno nuovo.',
    emailLabel: 'Indirizzo email',
    emailPlaceholder: "example{'@'}example.com",
    passwordLabel: 'Password',
    button: 'Login',
    forgotPassword: 'Password dimenticata?',
    invalidCredentials: 'Email o password non corretti.',
    genericError: 'Si è verificato un errore, riprova.',
    resetTitle: 'Recupera password',
    resetHint: "Inserisci l'email del tuo account: ti invieremo un link per impostare una nuova password.",
    resetButton: 'Invia link',
    resetSent: "Se l'email è registrata, riceverai a breve un link per reimpostare la password.",
    backToLogin: 'Torna al login',
    registerTitle: 'Crea un account',
    registerButton: 'Registrati',
    confirmPasswordLabel: 'Conferma password',
    passwordMismatch: 'Le password non coincidono.',
    emailInvalid: 'Inserisci un indirizzo email valido.',
    emailUnavailable: 'Questa email è già registrata.',
    emailDomainInvalid: 'Il dominio di questa email non riceve posta.',
    registerSuccess: 'Account creato! Controlla la tua email per confermarlo.',
    languageLabel: 'La tua lingua',
    statusLabel: 'Status',
    nicknameLabel: 'Nickname (pubblico)',
    nameLabel: 'Nome (privato)',
    phoneLabel: 'Numero di telefono (privato)',
    countryLabel: 'Paese (privato)',
    addressLabel: 'Indirizzo (privato)',
    optional: '- Facoltativo',
    newsletter: 'Tienimi aggiornato sulle novità.',
    acceptTerms: 'Accetto i {terms}',
    termsLink: 'termini e le condizioni di utilizzo',
    termsRequired: 'Devi accettare i termini e le condizioni di utilizzo.',
    statusRequired: 'Scegli il tuo status.',
    phoneInvalid: 'Numero di telefono non valido.',
    broadcasterHint: "Il profilo broadcaster viene abilitato dopo l'approvazione di un amministratore.",
    memberTypes: {
      visitor: 'Visitatore',
      broadcaster: 'Broadcaster',
      affiliate: 'Affiliato',
      advertiser: 'Inserzionista',
    },
  },
  terms: {
    title: 'Termini e condizioni',
    subtitle: 'Termini e condizioni di utilizzo del sito.',
    body: 'Testo dei termini e delle condizioni di utilizzo da definire.',
    updated: 'Ultimo aggiornamento: da definire.',
  },
};

const fr: MessageSchema = {
  common: {
    backHome: "Retour à l'accueil",
    dashboard: 'Tableau de bord',
    online: 'En ligne',
    loading: 'Chargement…',
  },
  menu: {
    home: 'Accueil',
    one: 'ONE',
    two: 'TWO',
    three: 'THREE',
    webcam: 'Webcam',
    contact: 'Contact',
  },
  actions: {
    advertise: 'Publicité',
    login: 'Connexion',
    goToDashboard: 'Accéder au tableau de bord',
  },
  home: {
    title: 'PublicSite',
    subtitle: 'Page publique statique (SSG) en architecture hybride.',
  },
  one: {
    title: 'ONE',
    subtitle: 'Section exclusive ONE du portail public.',
  },
  two: {
    title: 'TWO',
    subtitle: 'Section TWO du portail public.',
  },
  three: {
    title: 'THREE',
    subtitle: 'Section THREE du portail public.',
  },
  webcam: {
    title: 'Webcam',
    subtitle: 'Section Live Streaming et webcams interactives.',
    status: 'Canal de streaming prêt',
  },
  contact: {
    title: 'Contact',
    subtitle: 'Contactez notre équipe de support.',
    supportText: 'Assistance disponible 24/7',
  },
  advertise: {
    title: 'Publicité',
    subtitle: 'Faites la promotion de votre entreprise et gagnez en visibilité.',
    feature1: 'Bannières aux emplacements premium (Haut & Barre latérale)',
    feature2: 'Statistiques détaillées sur les clics et impressions',
    startNow: 'Commencer',
  },
  login: {
    title: 'Connexion',
    subtitle: 'Connectez-vous à votre compte ou créez-en un nouveau.',
    emailLabel: 'Adresse e-mail',
    emailPlaceholder: "example{'@'}example.com",
    passwordLabel: 'Mot de passe',
    button: 'Connexion',
    forgotPassword: 'Mot de passe oublié ?',
    invalidCredentials: 'E-mail ou mot de passe incorrect.',
    genericError: 'Une erreur est survenue, veuillez réessayer.',
    resetTitle: 'Récupérer le mot de passe',
    resetHint: "Saisissez l'e-mail de votre compte : nous vous enverrons un lien pour définir un nouveau mot de passe.",
    resetButton: 'Envoyer le lien',
    resetSent: "Si l'e-mail est enregistré, vous recevrez bientôt un lien pour réinitialiser le mot de passe.",
    backToLogin: 'Retour à la connexion',
    registerTitle: 'Créer un compte',
    registerButton: "S'inscrire",
    confirmPasswordLabel: 'Confirmer le mot de passe',
    passwordMismatch: 'Les mots de passe ne correspondent pas.',
    emailInvalid: 'Saisissez une adresse e-mail valide.',
    emailUnavailable: 'Cet e-mail est déjà enregistré.',
    emailDomainInvalid: 'Le domaine de cet e-mail ne reçoit pas de courrier.',
    registerSuccess: 'Compte créé ! Vérifiez votre e-mail pour le confirmer.',
    languageLabel: 'Votre langue',
    statusLabel: 'Statut',
    nicknameLabel: 'Pseudo (public)',
    nameLabel: 'Nom (privé)',
    phoneLabel: 'Numéro de téléphone (privé)',
    countryLabel: 'Pays (privé)',
    addressLabel: 'Adresse (privée)',
    optional: '- Facultatif',
    newsletter: 'Tenez-moi informé des nouveautés.',
    acceptTerms: "J'accepte les {terms}",
    termsLink: "conditions générales d'utilisation",
    termsRequired: "Vous devez accepter les conditions générales d'utilisation.",
    statusRequired: 'Choisissez votre statut.',
    phoneInvalid: 'Numéro de téléphone non valide.',
    broadcasterHint: "Le profil broadcaster est activé après l'approbation d'un administrateur.",
    memberTypes: {
      visitor: 'Visiteur',
      broadcaster: 'Broadcaster',
      affiliate: 'Affilié',
      advertiser: 'Annonceur',
    },
  },
  terms: {
    title: 'Conditions générales',
    subtitle: "Conditions générales d'utilisation du site.",
    body: "Texte des conditions générales d'utilisation à définir.",
    updated: 'Dernière mise à jour : à définir.',
  },
};

const de: MessageSchema = {
  common: {
    backHome: 'Zurück zur Startseite',
    dashboard: 'Dashboard',
    online: 'Online',
    loading: 'Wird geladen…',
  },
  menu: {
    home: 'Startseite',
    one: 'ONE',
    two: 'TWO',
    three: 'THREE',
    webcam: 'Webcam',
    contact: 'Kontakt',
  },
  actions: {
    advertise: 'Werbung',
    login: 'Anmelden',
    goToDashboard: 'Zum Dashboard',
  },
  home: {
    title: 'PublicSite',
    subtitle: 'Öffentliche statische Seite (SSG) in hybrider Architektur.',
  },
  one: {
    title: 'ONE',
    subtitle: 'Exklusiver ONE-Bereich des öffentlichen Portals.',
  },
  two: {
    title: 'TWO',
    subtitle: 'TWO-Bereich des öffentlichen Portals.',
  },
  three: {
    title: 'THREE',
    subtitle: 'THREE-Bereich des öffentlichen Portals.',
  },
  webcam: {
    title: 'Webcam',
    subtitle: 'Live-Streaming- und interaktive Webcam-Zone.',
    status: 'Streaming-Kanal bereit',
  },
  contact: {
    title: 'Kontakt',
    subtitle: 'Nehmen Sie Kontakt mit unserem Support-Team auf.',
    supportText: '24/7 Support verfügbar',
  },
  advertise: {
    title: 'Werbung',
    subtitle: 'Bewerben Sie Ihr Unternehmen und steigern Sie Ihre Reichweite.',
    feature1: 'Bannerplatzierungen mit hoher Reichweite (Top & Sidebar)',
    feature2: 'Detaillierte Klick- und Aufrufstatistiken',
    startNow: 'Jetzt starten',
  },
  login: {
    title: 'Anmeldung',
    subtitle: 'Melden Sie sich an oder erstellen Sie ein neues Konto.',
    emailLabel: 'E-Mail-Adresse',
    emailPlaceholder: "example{'@'}example.com",
    passwordLabel: 'Passwort',
    button: 'Anmelden',
    forgotPassword: 'Passwort vergessen?',
    invalidCredentials: 'E-Mail oder Passwort ist falsch.',
    genericError: 'Ein Fehler ist aufgetreten, bitte erneut versuchen.',
    resetTitle: 'Passwort wiederherstellen',
    resetHint: 'Geben Sie die E-Mail Ihres Kontos ein: Wir senden Ihnen einen Link, um ein neues Passwort festzulegen.',
    resetButton: 'Link senden',
    resetSent: 'Wenn die E-Mail registriert ist, erhalten Sie in Kürze einen Link zum Zurücksetzen des Passworts.',
    backToLogin: 'Zurück zur Anmeldung',
    registerTitle: 'Konto erstellen',
    registerButton: 'Registrieren',
    confirmPasswordLabel: 'Passwort bestätigen',
    passwordMismatch: 'Die Passwörter stimmen nicht überein.',
    emailInvalid: 'Geben Sie eine gültige E-Mail-Adresse ein.',
    emailUnavailable: 'Diese E-Mail ist bereits registriert.',
    emailDomainInvalid: 'Die Domain dieser E-Mail empfängt keine Post.',
    registerSuccess: 'Konto erstellt! Prüfen Sie Ihre E-Mail, um es zu bestätigen.',
    languageLabel: 'Ihre Sprache',
    statusLabel: 'Status',
    nicknameLabel: 'Nickname (öffentlich)',
    nameLabel: 'Name (privat)',
    phoneLabel: 'Telefonnummer (privat)',
    countryLabel: 'Land (privat)',
    addressLabel: 'Adresse (privat)',
    optional: '- Optional',
    newsletter: 'Halten Sie mich über Neuigkeiten auf dem Laufenden.',
    acceptTerms: 'Ich akzeptiere die {terms}',
    termsLink: 'Nutzungsbedingungen',
    termsRequired: 'Sie müssen die Nutzungsbedingungen akzeptieren.',
    statusRequired: 'Wählen Sie Ihren Status.',
    phoneInvalid: 'Ungültige Telefonnummer.',
    broadcasterHint: 'Das Broadcaster-Profil wird nach Freigabe durch einen Administrator aktiviert.',
    memberTypes: {
      visitor: 'Besucher',
      broadcaster: 'Broadcaster',
      affiliate: 'Affiliate',
      advertiser: 'Werbetreibender',
    },
  },
  terms: {
    title: 'Nutzungsbedingungen',
    subtitle: 'Nutzungsbedingungen der Website.',
    body: 'Text der Nutzungsbedingungen folgt.',
    updated: 'Letzte Aktualisierung: folgt.',
  },
};

const en: MessageSchema = {
  common: {
    backHome: 'Back to Home',
    dashboard: 'Dashboard',
    online: 'Online',
    loading: 'Loading…',
  },
  menu: {
    home: 'Home',
    one: 'ONE',
    two: 'TWO',
    three: 'THREE',
    webcam: 'Webcam',
    contact: 'Contact',
  },
  actions: {
    advertise: 'Advertise',
    login: 'Login',
    goToDashboard: 'Go to Dashboard',
  },
  home: {
    title: 'PublicSite',
    subtitle: 'Public static page (SSG) in hybrid architecture.',
  },
  one: {
    title: 'ONE',
    subtitle: 'Exclusive ONE section of the public portal.',
  },
  two: {
    title: 'TWO',
    subtitle: 'TWO section of the public portal.',
  },
  three: {
    title: 'THREE',
    subtitle: 'THREE section of the public portal.',
  },
  webcam: {
    title: 'Webcam',
    subtitle: 'Live Streaming and interactive webcam section.',
    status: 'Streaming Channel Ready',
  },
  contact: {
    title: 'Contact',
    subtitle: 'Get in touch with our customer support team.',
    supportText: '24/7 dedicated support',
  },
  advertise: {
    title: 'Advertise',
    subtitle: 'Promote your business and increase your visibility.',
    feature1: 'High-performing banner placements (Top & Sidebar)',
    feature2: 'Detailed click and view statistics',
    startNow: 'Start Now',
  },
  login: {
    title: 'Login',
    subtitle: 'Log in to your account or create a new one.',
    emailLabel: 'Email address',
    emailPlaceholder: "example{'@'}example.com",
    passwordLabel: 'Password',
    button: 'Login',
    forgotPassword: 'Forgotten your password?',
    invalidCredentials: 'Wrong email or password.',
    genericError: 'Something went wrong, please try again.',
    resetTitle: 'Recover password',
    resetHint: "Enter your account's email: we will send you a link to set a new password.",
    resetButton: 'Send link',
    resetSent: 'If the email is registered, you will shortly receive a link to reset your password.',
    backToLogin: 'Back to login',
    registerTitle: 'Create an account',
    registerButton: 'Register',
    confirmPasswordLabel: 'Confirm password',
    passwordMismatch: 'Passwords do not match.',
    emailInvalid: 'Enter a valid email address.',
    emailUnavailable: 'This email is already registered.',
    emailDomainInvalid: "This email's domain does not receive mail.",
    registerSuccess: 'Account created! Check your email to confirm it.',
    languageLabel: 'Your language',
    statusLabel: 'Status',
    nicknameLabel: 'Nickname (Public)',
    nameLabel: 'Name (Private)',
    phoneLabel: 'Phone Number (Private)',
    countryLabel: 'Country (Private)',
    addressLabel: 'Address (Private)',
    optional: '- Optional',
    newsletter: "Keep me posted on what's new.",
    acceptTerms: 'I accept the {terms}',
    termsLink: 'terms and conditions of use',
    termsRequired: 'You must accept the terms and conditions of use.',
    statusRequired: 'Choose your status.',
    phoneInvalid: 'Invalid phone number.',
    broadcasterHint: 'The broadcaster profile is enabled after an administrator approves it.',
    memberTypes: {
      visitor: 'Visitor',
      broadcaster: 'Broadcaster',
      affiliate: 'Affiliate',
      advertiser: 'Advertiser',
    },
  },
  terms: {
    title: 'Terms and conditions',
    subtitle: 'Terms and conditions of use of the site.',
    body: 'Text of the terms and conditions of use to be defined.',
    updated: 'Last updated: to be defined.',
  },
};

export default {
  it,
  fr,
  de,
  en,
};
