import "./geoJsonMap.scss";
import { clientOnly } from "@solidjs/start";
import type { GeoJsonMapProps } from "./types";

// Leaflet needs `window`, so the map itself only loads in the browser.
const GeoJsonMapClient = clientOnly(() => import("./GeoJsonMapClient"));

/**
 * Pan/zoom view of a GeoJSON floor plan (or any non-geographic shapes), with
 * optional drawing and editing. Size it with `--geo-map-height`.
 */
export function GeoJsonMap(props: GeoJsonMapProps) {
  return (
    <div class={`geo-map ${props.class ?? ""}`}>
      <GeoJsonMapClient {...props} fallback={<div class="geo-map__loading">Loading map…</div>} />
    </div>
  );
}
