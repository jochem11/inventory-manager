import type { Feature, FeatureCollection, GeoJsonProperties } from "geojson";

/** A point in plan units (e.g. metres), `[x, y]` with y pointing up. */
export type PlanPoint = [x: number, y: number];

export type GeoJsonMapProps = {
  /** The shapes to draw. Coordinates are plain `[x, y]` plan units, not lat/lng. */
  data: FeatureCollection;
  /** Feature `id` to highlight. */
  selectedId?: string | number;
  /** Shows the draw/edit/drag/remove toolbar. */
  editable?: boolean;
  onFeatureClick?: (feature: Feature) => void;
  /** Clicked on the map but not on a shape. */
  onBackgroundClick?: () => void;
  /** Fires with the full updated collection after any draw, edit, drag or removal. */
  onChange?: (data: FeatureCollection) => void;
  /** Fires with the newly drawn feature (after `onChange`). */
  onCreate?: (feature: Feature) => void;
  /** Properties given to newly drawn shapes. */
  newFeatureProperties?: () => GeoJsonProperties;
  /** Tooltip / label text; defaults to `properties.name`. */
  label?: (feature: Feature) => string;
  /**
   * Print labels on the shapes instead of only showing them on hover. Pass a
   * function to label some features (e.g. rooms) and keep hover for the rest.
   */
  showLabels?: boolean | ((feature: Feature) => boolean);
  /** A scanned floor plan to trace over, stretched between two corners. */
  backgroundImage?: { url: string; bounds: [PlanPoint, PlanPoint] };
  class?: string;
};
