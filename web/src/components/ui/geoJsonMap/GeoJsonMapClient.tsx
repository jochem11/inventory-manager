import "leaflet/dist/leaflet.css";
import "@geoman-io/leaflet-geoman-free/dist/leaflet-geoman.css";
import L from "./leaflet";
import "@geoman-io/leaflet-geoman-free";
import type { Feature, FeatureCollection } from "geojson";
import { createEffect, onCleanup, onMount } from "solid-js";
import type { GeoJsonMapProps, PlanPoint } from "./types";

type FeatureLayer = L.Path & { feature: Feature };

const SHAPE_CLASS = "geo-map__shape";

// GeoJSON is [x, y]; Leaflet's LatLng is (lat, lng) = (y, x) in CRS.Simple.
const toLatLng = ([x, y]: PlanPoint) => L.latLng(y, x);

/**
 * The Leaflet side of `GeoJsonMap`. Browser-only (Leaflet touches `window` on
 * import), so it's only ever loaded through `clientOnly`.
 */
export default function GeoJsonMapClient(props: GeoJsonMapProps) {
  let container!: HTMLDivElement;
  let map!: L.Map;
  let shapes!: L.GeoJSON;
  let image: L.ImageOverlay | undefined;
  // The collection we last reported via `onChange`. When it comes straight
  // back as `data`, the map already shows it, so we skip a full rebuild
  // (which would also interrupt an edit in progress).
  let lastEmitted: FeatureCollection | undefined;
  let fitted = false;

  const labelOf = (feature: Feature) =>
    props.label?.(feature) ?? String(feature.properties?.name ?? "");

  const kindClass = (feature: Feature) => {
    const kind = feature.properties?.kind;
    return typeof kind === "string" ? ` ${SHAPE_CLASS}--${kind}` : "";
  };

  const emit = () => {
    lastEmitted = shapes.toGeoJSON() as FeatureCollection;
    props.onChange?.(lastEmitted);
    fitLabels();
  };

  const highlight = () => {
    shapes.eachLayer((layer) => {
      const { feature } = layer as FeatureLayer;
      (layer as L.Path)
        .getElement()
        ?.classList.toggle(`${SHAPE_CLASS}--selected`, feature.id === props.selectedId);
    });
  };

  const hasLabel = (feature: Feature) =>
    typeof props.showLabels === "function" ? props.showLabels(feature) : !!props.showLabels;

  // Hide printed labels that don't fit inside their shape at this zoom
  // level (they'd spill over neighbours); zooming in brings them back.
  const fitLabels = () => {
    shapes.eachLayer((layer) => {
      const label = layer.getTooltip()?.getElement();
      const shape = (layer as L.Path).getElement();
      if (!label || !shape || !label.classList.contains("geo-map__label")) return;
      label.classList.remove("geo-map__label--hidden");
      const room = shape.getBoundingClientRect();
      const fits = label.offsetWidth <= room.width - 4 && label.offsetHeight <= room.height - 4;
      label.classList.toggle("geo-map__label--hidden", !fits);
    });
  };

  const wireFeature = (feature: Feature, layer: L.Layer) => {
    layer.bindTooltip(
      () => labelOf(feature),
      hasLabel(feature)
        ? { permanent: true, direction: "center", className: "geo-map__label" }
        : { sticky: true, className: "geo-map__tooltip" },
    );
    layer.on("click", (e) => {
      // Keep the map's own click (= "clicked background") from firing too.
      L.DomEvent.stopPropagation(e);
      props.onFeatureClick?.(feature);
    });
  };

  onMount(() => {
    map = L.map(container, {
      crs: L.CRS.Simple,
      minZoom: -2,
      maxZoom: 7,
      zoomSnap: 0.25,
      attributionControl: false,
    });

    shapes = L.geoJSON(undefined, {
      style: (feature) => ({ className: `${SHAPE_CLASS}${feature ? kindClass(feature) : ""}` }),
      onEachFeature: wireFeature,
    }).addTo(map);

    map.on("click", () => props.onBackgroundClick?.());
    map.on("zoomend", fitLabels);

    // --- Drawing & editing (Geoman) ---------------------------------------
    map.pm.setGlobalOptions({ snappable: true, snapDistance: 12, allowSelfIntersection: false });
    map.pm.setPathOptions({ className: SHAPE_CLASS });

    map.on("pm:create", (e) => {
      const layer = e.layer as L.Polygon;
      const feature: Feature = {
        type: "Feature",
        id: crypto.randomUUID(),
        properties: props.newFeatureProperties?.() ?? { name: "New shape" },
        geometry: (layer.toGeoJSON() as Feature).geometry,
      };
      // Hand the drawn layer over to our GeoJSON group so it's styled,
      // clickable and included in `toGeoJSON()` like the others.
      map.removeLayer(layer);
      shapes.addData(feature);
      emit();
      props.onCreate?.(feature);
      highlight();
    });

    map.on("pm:remove", (e) => {
      shapes.removeLayer(e.layer);
      emit();
    });

    // Child layer events bubble up to the GeoJSON group.
    shapes.on("pm:edit pm:dragend", emit);

    // Leaflet caches its size; keep it right when the sidebar collapses etc.
    const resize = new ResizeObserver(() => {
      map.invalidateSize();
      fitLabels();
    });
    resize.observe(container);

    onCleanup(() => {
      resize.disconnect();
      map.remove();
    });

    // --- Reactive props ---------------------------------------------------
    createEffect(() => {
      const data = props.data;
      if (data === lastEmitted) return;
      shapes.clearLayers();
      shapes.addData(data);
      highlight();
      fitLabels();

      // Frame the plan once; later changes keep the user's view.
      if (!fitted && shapes.getLayers().length > 0) {
        map.fitBounds(shapes.getBounds(), { padding: [24, 24] });
        fitted = true;
        fitLabels();
      } else if (!fitted) {
        map.setView([0, 0], 0);
      }
    });

    createEffect(highlight);

    createEffect(() => {
      image?.remove();
      image = undefined;
      const bg = props.backgroundImage;
      if (!bg) return;
      image = L.imageOverlay(bg.url, L.latLngBounds(bg.bounds.map(toLatLng)), {
        className: "geo-map__image",
      })
        .addTo(map)
        .bringToBack();
    });

    createEffect(() => {
      if (props.editable) {
        map.pm.addControls({
          position: "topleft",
          drawPolygon: true,
          drawRectangle: true,
          editMode: true,
          dragMode: true,
          removalMode: true,
          drawMarker: false,
          drawCircleMarker: false,
          drawPolyline: false,
          drawCircle: false,
          drawText: false,
          cutPolygon: false,
          rotateMode: false,
        });
      } else {
        map.pm.disableDraw();
        map.pm.disableGlobalEditMode();
        map.pm.disableGlobalDragMode();
        map.pm.disableGlobalRemovalMode();
        map.pm.removeControls();
      }
    });
  });

  return <div ref={container} class="geo-map__canvas" />;
}
