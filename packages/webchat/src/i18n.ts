/** Testi del pacchetto, da unire ai messaggi di ogni app sotto la chiave 'webchat'. */
export interface WebchatMessages {
  camera: string;
  screen: string;
  waiting: string;
  resize: string;
  chatTitle: string;
  chatEmpty: string;
  chatPlaceholder: string;
  chatSend: string;
  scrollBottom: string;
  readOnly: string;
  user: string;
  cancel: string;
    webcam: {
      title: string;
      presetsLabel: string;
      presetsHelp: string;
      presets: { balanced: string; high: string; low: string; custom: string };
      device: string;
      defaultDevice: string;
      deviceHelp: string;
      resolution: string;
      resolutionHelp: string;
      width: string;
      height: string;
      heightCalculated: string;
      heightCalculatedHelp: string;
      aspectRatio: string;
      aspectRatioHelp: string;
      frameRate: string;
      frameRateHelp: string;
      bitrate: string;
      bitrateHelp: string;
      codec: string;
      codecHelp: string;
      simulcast: string;
      simulcastHelp: string;
      degradation: string;
      degradationHelp: string;
      degradationOptions: { maintainFramerate: string; maintainResolution: string; balanced: string };
      savedSuccess: string;
      apply: string;
      camera: string;
      open: string;
    };
  }
  
  export const webchatMessages: Record<'en' | 'it' | 'fr' | 'de' | 'es', WebchatMessages> = {
    en: {
      camera: 'Camera',
      screen: 'Screen',
      waiting: 'Waiting for the broadcast to start.',
      resize: 'Drag to resize',
      chatTitle: 'Live chat',
      chatEmpty: 'No messages yet.',
      chatPlaceholder: 'Write a message…',
      chatSend: 'Send',
      scrollBottom: 'Go to the latest messages',
      readOnly: 'Log in to write in the chat.',
      user: 'User',
      cancel: 'Cancel',
    webcam: {
      title: 'Webcam & Streaming Configuration',
      presetsLabel: 'Quality Preset',
      presetsHelp: 'Select an optimized profile or choose Custom to adjust each parameter individually.',
      presets: {
        balanced: 'HD 720p - Balanced (Recommended)',
        high: 'Full HD 1080p - High Quality',
        low: 'SD 480p - Data Saver / Low Bandwidth',
        custom: 'Custom',
      },
      device: 'Camera Device',
      defaultDevice: 'System default',
      deviceHelp: 'Select which webcam or capture card to use among the connected video devices.',
      resolution: 'Resolution (Dimensions)',
      resolutionHelp: 'Width and height in pixels. Higher resolutions (e.g. 1080p) offer sharper video but require higher CPU and upload bandwidth.',
      width: 'Width (px)',
      height: 'Height (px)',
      heightCalculated: 'Height (px) - Auto-calculated',
      heightCalculatedHelp: 'Height is automatically determined by the selected width and aspect ratio (e.g. 16:9, 4:3) to preserve perfect capture proportions without distortion.',
      aspectRatio: 'Aspect Ratio',
      aspectRatioHelp: 'Ratio between video width and height. 16:9 is standard widescreen, 4:3 is traditional, 9:16 is vertical mobile.',
      frameRate: 'Frame Rate (FPS)',
      frameRateHelp: 'Frames per second transmitted. 30 FPS is standard for smooth motion; 20-24 FPS saves bandwidth; 60 FPS offers ultra-fluid motion.',
      bitrate: 'Transfer Rate / Max Bitrate (kbps)',
      bitrateHelp: 'Maximum data sent per second in kilobits (1700 kbps = 1.7 Mbps). Higher values improve visual clarity but require a faster upload connection.',
      codec: 'Video Codec',
      codecHelp: 'Compression algorithm. VP8 provides maximum compatibility; H264 uses hardware acceleration; VP9 and AV1 provide higher compression efficiency.',
      simulcast: 'Simulcast (Multi-quality)',
      simulcastHelp: 'When enabled, publishes full resolution alongside scaled-down layers so viewers with slow connections can watch smoothly.',
      degradation: 'Network Degradation Strategy',
      degradationHelp: "Defines behavior during upload bandwidth drops: 'Maintain framerate' drops resolution to keep video fluid; 'Maintain resolution' drops frames to keep sharpness.",
      degradationOptions: {
        maintainFramerate: 'Maintain framerate (FPS prioritized)',
        maintainResolution: 'Maintain resolution (sharpness prioritized)',
        balanced: 'Balanced',
      },
      savedSuccess: 'Webcam settings saved and applied successfully.',
      apply: 'Save and Apply',
      camera: 'Webcam',
      open: 'Webcam settings',
    },
  },
  it: {
    camera: 'Camera',
    screen: 'Schermo',
    waiting: 'In attesa che la diretta inizi.',
    resize: 'Trascina per ridimensionare',
    chatTitle: 'Chat dal vivo',
    chatEmpty: 'Ancora nessun messaggio.',
    chatPlaceholder: 'Scrivi un messaggio…',
    chatSend: 'Invia',
    scrollBottom: 'Vai agli ultimi messaggi',
    readOnly: 'Accedi per scrivere in chat.',
    user: 'Utente',
    cancel: 'Annulla',
    webcam: {
      title: 'Configurazione Webcam e Streaming',
      presetsLabel: 'Preset Qualità',
      presetsHelp: 'Seleziona un profilo predefinito ottimizzato oppure scegli Personalizzato per modificare ogni parametro a mano.',
      presets: {
        balanced: 'HD 720p - Bilanciato (Consigliato)',
        high: 'Full HD 1080p - Alta Qualità',
        low: 'SD 480p - Risparmio Dati / Banda Bassa',
        custom: 'Personalizzato',
      },
      device: 'Dispositivo Videocamera',
      defaultDevice: 'Predefinito di sistema',
      deviceHelp: 'Permette di selezionare quale webcam o scheda di acquisizione utilizzare tra quelle connesse al computer.',
      resolution: 'Risoluzione (Dimensioni)',
      resolutionHelp: 'Larghezza e altezza del fotogramma in pixel. Risoluzioni maggiori (es. 1080p) offrono più dettaglio ma richiedono più CPU e banda in upload.',
      width: 'Larghezza (px)',
      height: 'Altezza (px)',
      heightCalculated: 'Altezza (px) - Calcolata automaticamente',
      heightCalculatedHelp: "L'altezza viene calcolata automaticamente in base alla larghezza e alla proporzione selezionata (es. 16:9, 4:3) per garantire sempre proporzioni perfette senza distorsioni.",
      aspectRatio: 'Proporzioni (Aspect Ratio)',
      aspectRatioHelp: 'Rapporto tra larghezza e altezza del frame. 16:9 è il formato widescreen standard, 4:3 è il formato webcam tradizionale, 9:16 per streaming verticale.',
      frameRate: 'Frequenza Fotogrammi (FPS)',
      frameRateHelp: 'Numero di immagini trasmesse al secondo. 30 FPS è lo standard per un movimento naturale; 20-24 FPS risparmia banda; 60 FPS per massima fluidità.',
      bitrate: 'Transfer Rate / Bitrate Massimo (kbps)',
      bitrateHelp: 'Tetto massimo di dati inviati al secondo in kilobit (1700 kbps = 1.7 Mbps). Un valore più alto aumenta la nitidezza ma richiede più velocità di connessione in upload.',
      codec: 'Codec Video',
      codecHelp: "Algoritmo di compressione video. VP8 offre massima compatibilità; H264 usa l'accelerazione hardware; VP9 e AV1 offrono maggiore efficienza sui browser moderni.",
      simulcast: 'Simulcast (Multi-qualità)',
      simulcastHelp: 'Se abilitato, invia contemporaneamente lo stream a risoluzione piena e versioni ridotte. Gli utenti con connessione lenta vedranno la versione ridotta senza blocchi.',
      degradation: 'Strategia di Degradazione Rete',
      degradationHelp: "Comportamento quando la connessione cala: 'Mantieni fluidità' abbassa la risoluzione ma preserva i fotogrammi; 'Mantieni risoluzione' riduce i fotogrammi preservando la nitidezza.",
      degradationOptions: {
        maintainFramerate: 'Mantieni fluidità (FPS prioritario)',
        maintainResolution: 'Mantieni risoluzione (nitidezza prioritaria)',
        balanced: 'Bilanciata',
      },
      savedSuccess: 'Impostazioni webcam salvate e applicate con successo.',
      apply: 'Salva e Applica',
      camera: 'Webcam',
      open: 'Impostazioni webcam',
    },
  },
  fr: {
    camera: 'Caméra',
    screen: 'Écran',
    waiting: 'En attente du début du direct.',
    resize: 'Faites glisser pour redimensionner',
    chatTitle: 'Chat en direct',
    chatEmpty: 'Aucun message pour le moment.',
    chatPlaceholder: 'Écrivez un message…',
    chatSend: 'Envoyer',
    scrollBottom: 'Aller aux derniers messages',
    readOnly: 'Connectez-vous pour écrire dans le chat.',
    user: 'Utilisateur',
    cancel: 'Annuler',
    webcam: {
      title: 'Configuration de la webcam et du streaming',
      presetsLabel: 'Préréglage de qualité',
      presetsHelp: 'Sélectionnez un profil prédéfini optimisé ou choisissez Personnalisé pour modifier chaque paramètre.',
      presets: {
        balanced: 'HD 720p - Équilibré (Recommandé)',
        high: 'Full HD 1080p - Haute Qualité',
        low: 'SD 480p - Économie de données / Bas débit',
        custom: 'Personnalisé',
      },
      device: 'Périphérique de caméra',
      defaultDevice: 'Par défaut du système',
      deviceHelp: 'Permet de sélectionner la caméra ou carte de capture vidéo à utiliser parmi celles connectées.',
      resolution: 'Résolution (Dimensions)',
      resolutionHelp: "Largeur et hauteur en pixels. Les résolutions plus élevées (ex. 1080p) offrent plus de détails mais demandent plus de CPU et d'upload.",
      width: 'Largeur (px)',
      height: 'Hauteur (px)',
      heightCalculated: 'Hauteur (px) - Calculée automatiquement',
      heightCalculatedHelp: "La hauteur est automatiquement déterminée selon la largeur et le format d'image (ex. 16:9, 4:3) afin de préserver des proportions parfaites sans distorsion.",
      aspectRatio: "Format d'image (Aspect Ratio)",
      aspectRatioHelp: 'Rapport entre largeur et hauteur. 16:9 est le format large standard, 4:3 est classique, 9:16 est vertical pour mobile.',
      frameRate: "Fréquence d'images (FPS)",
      frameRateHelp: "Nombre d'images par seconde. 30 FPS est le standard pour un mouvement fluide; 20-24 FPS économise la bande passante; 60 FPS pour fluidité maximale.",
      bitrate: 'Débit binaire maximal (kbps)',
      bitrateHelp: "Débit maximal envoyé en kilobits par seconde (1700 kbps = 1.7 Mbps). Une valeur plus élevée améliore la clarté mais demande plus d'upload.",
      codec: 'Codec vidéo',
      codecHelp: "Algorithme de compression vidéo. VP8 assure la compatibilité; H264 utilise l'accélération matérielle; VP9 et AV1 offrent une compression supérieure.",
      simulcast: 'Simulcast (Multi-qualité)',
      simulcastHelp: 'Si activé, envoie le flux en haute résolution et des versions réduites pour les spectateurs avec une connexion faible.',
      degradation: 'Stratégie de dégradation réseau',
      degradationHelp: "Comportement en cas de baisse de débit: 'Maintenir la fluidité' réduit la résolution pour rester fluide; 'Maintenir la résolution' réduit les images pour garder la netteté.",
      degradationOptions: {
        maintainFramerate: 'Maintenir la fluidité (FPS prioritaire)',
        maintainResolution: 'Maintenir la résolution (netteté prioritaire)',
        balanced: 'Équilibrée',
      },
      savedSuccess: 'Paramètres de la webcam enregistrés et appliqués avec succès.',
      apply: 'Enregistrer et appliquer',
      camera: 'Webcam',
      open: 'Paramètres webcam',
    },
  },
  de: {
    camera: 'Kamera',
    screen: 'Bildschirm',
    waiting: 'Warten auf den Beginn der Übertragung.',
    resize: 'Zum Ändern der Größe ziehen',
    chatTitle: 'Live-Chat',
    chatEmpty: 'Noch keine Nachrichten.',
    chatPlaceholder: 'Nachricht schreiben…',
    chatSend: 'Senden',
    scrollBottom: 'Zu den neuesten Nachrichten',
    readOnly: 'Melde dich an, um im Chat zu schreiben.',
    user: 'Benutzer',
    cancel: 'Abbrechen',
    webcam: {
      title: 'Webcam- und Streaming-Konfiguration',
      presetsLabel: 'Qualitäts-Voreinstellung',
      presetsHelp: 'Wählen Sie ein optimiertes Profil oder wählen Sie Benutzerdefiniert, um jeden Parameter anzupassen.',
      presets: {
        balanced: 'HD 720p - Ausgewogen (Empfohlen)',
        high: 'Full HD 1080p - Hohe Qualität',
        low: 'SD 480p - Datensparmodus / Geringe Bandbreite',
        custom: 'Benutzerdefiniert',
      },
      device: 'Kameragerät',
      defaultDevice: 'Systemstandard',
      deviceHelp: 'Ermöglicht die Auswahl der zu verwendenden Webcam oder Erfassungskarte unter den angeschlossenen Geräten.',
      resolution: 'Auflösung (Dimensionen)',
      resolutionHelp: 'Breite und Höhe in Pixeln. Höhere Auflösungen (z. B. 1080p) bieten mehr Details, erfordern jedoch mehr CPU und Upload-Bandbreite.',
      width: 'Breite (px)',
      height: 'Höhe (px)',
      heightCalculated: 'Höhe (px) - Automatisch berechnet',
      heightCalculatedHelp: 'Die Höhe wird automatisch anhand der gewählten Breite und des Seitenverhältnisses (z. B. 16:9, 4:3) berechnet, um perfekte Proportionen ohne Verzerrung zu gewährleisten.',
      aspectRatio: 'Seitenverhältnis (Aspect Ratio)',
      aspectRatioHelp: 'Verhältnis von Breite zu Höhe. 16:9 ist Breitbildstandard, 4:3 traditionell, 9:16 vertikal für Mobilgeräte.',
      frameRate: 'Bildrate (FPS)',
      frameRateHelp: 'Übertragene Bilder pro Sekunde. 30 FPS ist Standard; 20-24 FPS spart Bandbreite; 60 FPS bietet maximale Flüssigkeit.',
      bitrate: 'Max. Transferrate / Bitrate (kbps)',
      bitrateHelp: 'Maximale Datenmenge pro Sekunde in Kilobit (1700 kbps = 1.7 Mbps). Höhere Werte verbessern die Bildschärfe, erfordern aber mehr Upload.',
      codec: 'Video-Codec',
      codecHelp: 'Komprimierungsalgorithmus. VP8 garantiert maximale Kompatibilität; H264 nutzt Hardwarebeschleunigung; VP9 und AV1 bieten bessere Effizienz.',
      simulcast: 'Simulcast (Multi-Qualität)',
      simulcastHelp: 'Wenn aktiviert, sendet der Streamer gleichzeitig volle Auflösung und reduzierte Versionen für langsamere Verbindungen.',
      degradation: 'Netzwerk-Degradierungsstrategie',
      degradationHelp: "Verhalten bei Bandbreitenabfall: 'Bildrate beibehalten' senkt Auflösung für Flüssigkeit; 'Auflösung beibehalten' senkt FPS für Schärfe.",
      degradationOptions: {
        maintainFramerate: 'Bildrate beibehalten (FPS priorisiert)',
        maintainResolution: 'Auflösung beibehalten (Schärfe priorisiert)',
        balanced: 'Ausgewogen',
      },
      savedSuccess: 'Webcam-Einstellungen erfolgreich gespeichert und angewendet.',
      apply: 'Speichern und anwenden',
      camera: 'Webcam',
      open: 'Webcam-Einstellungen',
    },
  },
  es: {
    camera: 'Cámara',
    screen: 'Pantalla',
    waiting: 'Esperando a que empiece la emisión.',
    resize: 'Arrastra para cambiar el tamaño',
    chatTitle: 'Chat en directo',
    chatEmpty: 'Todavía no hay mensajes.',
    chatPlaceholder: 'Escribe un mensaje…',
    chatSend: 'Enviar',
    scrollBottom: 'Ir a los últimos mensajes',
    readOnly: 'Inicia sesión para escribir en el chat.',
    user: 'Usuario',
    cancel: 'Cancelar',
    webcam: {
      title: 'Configuración de cámara web y transmisión',
      presetsLabel: 'Preajuste de calidad',
      presetsHelp: 'Selecciona un perfil predefinido optimizado o elige Personalizado para modificar cada parámetro manualmente.',
      presets: {
        balanced: 'HD 720p - Equilibrado (Recomendado)',
        high: 'Full HD 1080p - Alta Calidad',
        low: 'SD 480p - Ahorro de datos / Bajo ancho de banda',
        custom: 'Personalizado',
      },
      device: 'Dispositivo de cámara',
      defaultDevice: 'Predeterminado del sistema',
      deviceHelp: 'Permite seleccionar qué cámara web o capturadora de video utilizar entre las conectadas al equipo.',
      resolution: 'Resolución (Dimensiones)',
      resolutionHelp: 'Ancho y alto en píxeles. Resoluciones mayores (ej. 1080p) ofrecen más detalle pero consumen más CPU y ancho de banda de subida.',
      width: 'Ancho (px)',
      height: 'Alto (px)',
      heightCalculated: 'Altura (px) - Calculada automáticamente',
      heightCalculatedHelp: 'La altura se determina automáticamente según el ancho y la relación de aspecto (ej. 16:9, 4:3) para preservar proporciones perfectas sin distorsión.',
      aspectRatio: 'Relación de aspecto (Aspect Ratio)',
      aspectRatioHelp: 'Relación entre ancho y alto del video. 16:9 es panorámico estándar, 4:3 clásico, 9:16 vertical para móviles.',
      frameRate: 'Fotogramas por segundo (FPS)',
      frameRateHelp: 'Número de fotogramas transmitidos por segundo. 30 FPS es estándar para movimiento fluido; 20-24 FPS ahorra ancho de banda; 60 FPS para máxima fluidez.',
      bitrate: 'Tasa de transferencia / Bitrate máximo (kbps)',
      bitrateHelp: 'Cantidad máxima de datos enviados por segundo en kilobits (1700 kbps = 1.7 Mbps). Valores más altos reducen artefactos pero requieren más subida.',
      codec: 'Códec de video',
      codecHelp: 'Algoritmo de compresión de video. VP8 garantiza máxima compatibilidad; H264 usa aceleración por hardware; VP9 y AV1 ofrecen mayor compresión.',
      simulcast: 'Simulcast (Multi-calidad)',
      simulcastHelp: 'Si está activo, transmite la resolución completa junto con versiones reducidas para espectadores con conexiones lentas.',
      degradation: 'Estrategia de degradación de red',
      degradationHelp: "Comportamiento ante caídas de conexión: 'Mantener fluidez' reduce resolución manteniendo FPS; 'Mantener resolución' reduce FPS preservando nitidez.",
      degradationOptions: {
        maintainFramerate: 'Mantener fluidez (FPS prioritario)',
        maintainResolution: 'Mantener resolución (nitidez prioritaria)',
        balanced: 'Equilibrada',
      },
      savedSuccess: 'Configuración de la cámara web guardada y aplicada con éxito.',
      apply: 'Guardar y aplicar',
      camera: 'Cámara web',
      open: 'Ajustes de webcam',
    },
  },
};
