import L from "leaflet";

// Geoman doesn't import Leaflet; it patches the global `window.L`. Importing
// this module before Geoman guarantees the global exists when Geoman runs.
(window as unknown as { L: typeof L }).L = L;

export default L;
