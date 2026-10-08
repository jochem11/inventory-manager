import "./comingSoon.scss";
import { useLocation } from "@solidjs/router";
import { Icon, PageHeader } from "~/components/ui";
import { findNavTrail } from "~/constants/navigation";

/** Stand-in for navigation entries whose page hasn't been built yet. */
export default function ComingSoon() {
  const location = useLocation();
  const title = () => findNavTrail(location.pathname).at(-1)?.label ?? "Coming soon";

  return (
    <>
      <PageHeader title={title()} />
      <div class="coming-soon">
        <Icon name="package" class="coming-soon__icon" />
        <p class="coming-soon__title">Nothing here yet</p>
        <p class="coming-soon__text">This section is on its way.</p>
      </div>
    </>
  );
}
