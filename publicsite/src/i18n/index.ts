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
    button: string;
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
    subtitle: "Accedi all'area riservata o alla dashboard amministrativa.",
    button: 'Accedi alla Dashboard',
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
    subtitle: 'Accédez à votre espace réservé ou au tableau de bord.',
    button: 'Accéder au tableau de bord',
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
    subtitle: 'Melden Sie sich für den geschützten Bereich oder das Dashboard an.',
    button: 'Zum Dashboard anmelden',
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
    subtitle: 'Access the private area or administrative dashboard.',
    button: 'Access Dashboard',
  },
};

export default {
  it,
  fr,
  de,
  en,
};
