// htmx wird von Vite mitgebündelt, damit die App vollständig offline läuft.
// Eigene Logik gibt es hier bewusst nicht: alle Interaktionen laufen über
// hx-*-Attribute gegen die Go-Handler in app/.
import 'htmx.org';
import './style.css';
