/** Testi del pacchetto, da unire ai messaggi di ogni app sotto la chiave "webchat". */
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
  },
};
