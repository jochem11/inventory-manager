import "./locations.scss";
import type { Feature, FeatureCollection, Position } from "geojson";
import { createSignal, Show } from "solid-js";
import { ConfirmDialog, PromptDialog } from "~/components/dialogs";
import { Button, GeoJsonMap, Icon, PageHeader } from "~/components/ui";
import sampleFloorPlan from "~/data/floorPlan.json";
import { createDialog } from "~/hooks";

// Temporary demo data (metres, y pointing up) until plans come from a backend.
// JSON imports are typed loosely (`"type": string`), hence the cast.
const SAMPLE_FLOOR_PLAN = sampleFloorPlan as FeatureCollection;

type PlanKind = "room" | "area" | "shelf";

const KINDS: PlanKind[] = ["room", "area", "shelf"];
const KIND_LABEL: Record<PlanKind, string> = { room: "Room", area: "Area", shelf: "Shelf" };

const kindOf = (feature: Feature): PlanKind =>
  KINDS.find((k) => k === feature.properties?.kind) ?? "room";

// Shoelace formula; plan units are metres, so this is m².
const ringArea = (ring: Position[]) =>
  Math.abs(
    ring.reduce((sum, [x1, y1], i) => {
      const [x2, y2] = ring[(i + 1) % ring.length]!;
      return sum + (x1! * y2! - x2! * y1!);
    }, 0),
  ) / 2;

const areaOf = (feature: Feature) => {
  if (feature.geometry.type !== "Polygon") return undefined;
  const [outer, ...holes] = feature.geometry.coordinates;
  if (!outer) return undefined;
  return ringArea(outer) - holes.reduce((sum, hole) => sum + ringArea(hole), 0);
};

export default function Locations() {
  // A signal (not a store) so the map gets back the exact object it emitted
  // and can skip redrawing — see `lastEmitted` in GeoJsonMapClient.
  const [plan, setPlan] = createSignal<FeatureCollection>(SAMPLE_FLOOR_PLAN);
  const [selectedId, setSelectedId] = createSignal<string | number>();
  const [editing, setEditing] = createSignal(false);

  const confirm = createDialog(ConfirmDialog);
  const prompt = createDialog(PromptDialog);

  const selected = () => plan().features.find((f) => f.id === selectedId());
  const count = (kind: PlanKind) => plan().features.filter((f) => kindOf(f) === kind).length;

  const updateFeature = (id: string | number, update: (f: Feature) => Feature) =>
    setPlan((p) => ({ ...p, features: p.features.map((f) => (f.id === id ? update(f) : f)) }));

  const rename = async (feature: Feature) => {
    const name = await prompt.open({
      title: `Rename ${KIND_LABEL[kindOf(feature)].toLowerCase()}`,
      label: "Name",
      initialValue: String(feature.properties?.name ?? ""),
    });
    if (name) updateFeature(feature.id!, (f) => ({ ...f, properties: { ...f.properties, name } }));
  };

  const setKind = (feature: Feature, kind: PlanKind) =>
    updateFeature(feature.id!, (f) => ({ ...f, properties: { ...f.properties, kind } }));

  const remove = async (feature: Feature) => {
    const ok = await confirm.open({
      title: `Delete "${feature.properties?.name}"?`,
      message: "The shape will be removed from the floor plan.",
      confirmLabel: "Delete",
      variant: "danger",
    });
    if (!ok) return;
    setPlan((p) => ({ ...p, features: p.features.filter((f) => f.id !== feature.id) }));
    setSelectedId(undefined);
  };

  return (
    <>
      <PageHeader
        title="Locations"
        description={`${count("room")} rooms · ${count("shelf")} shelves & cabinets`}
        actions={
          <Button
            variant={editing() ? "primary" : "outline"}
            startIcon={<Icon name={editing() ? "x" : "sliders"} />}
            onClick={() => setEditing((e) => !e)}>
            {editing() ? "Done editing" : "Edit layout"}
          </Button>
        }
      />

      <div class="locations">
        <GeoJsonMap
          class="locations__map"
          data={plan()}
          selectedId={selectedId()}
          editable={editing()}
          onChange={setPlan}
          onFeatureClick={(f) => setSelectedId(f.id)}
          onBackgroundClick={() => setSelectedId(undefined)}
          onCreate={(f) => setSelectedId(f.id)}
          newFeatureProperties={() => ({ name: "New room", kind: "room" })}
          // Name rooms and corridors on the plan; furniture shows on hover.
          showLabels={(f) => kindOf(f) !== "shelf"}
        />

        <aside class="locations__panel" aria-live="polite">
          <Show
            when={selected()}
            fallback={
              <div class="locations__hint">
                <Icon name="map-pin" class="locations__hint-icon" />
                <p>
                  {editing()
                    ? "Use the toolbar on the map to draw, reshape, move or delete rooms."
                    : "Click a room or shelf on the plan to see its details."}
                </p>
              </div>
            }>
            {(feature) => (
              <>
                <div class="locations__heading">
                  <span class={`locations__badge locations__badge--${kindOf(feature())}`}>
                    {KIND_LABEL[kindOf(feature())]}
                  </span>
                  <h2 class="locations__name">{String(feature().properties?.name ?? "Unnamed")}</h2>
                </div>

                <dl class="locations__facts">
                  <Show when={areaOf(feature())}>
                    {(area) => (
                      <>
                        <dt>Area</dt>
                        <dd>{area().toFixed(1)} m²</dd>
                      </>
                    )}
                  </Show>
                  <dt>Type</dt>
                  <dd>
                    <div class="locations__kind" role="group" aria-label="Type">
                      {KINDS.map((kind) => (
                        <Button
                          size="sm"
                          variant={kindOf(feature()) === kind ? "secondary" : "ghost"}
                          aria-pressed={kindOf(feature()) === kind}
                          onClick={() => setKind(feature(), kind)}>
                          {KIND_LABEL[kind]}
                        </Button>
                      ))}
                    </div>
                  </dd>
                </dl>

                <div class="locations__actions">
                  <Button variant="outline" onClick={() => rename(feature())}>
                    Rename
                  </Button>
                  <Button variant="ghost" class="locations__delete" onClick={() => remove(feature())}>
                    Delete
                  </Button>
                </div>
              </>
            )}
          </Show>
        </aside>
      </div>
    </>
  );
}
