import { A } from "@solidjs/router";
import { HttpStatusCode } from "@solidjs/start";
import { PageHeader } from "~/components/ui";

export default function NotFound() {
  return (
    <>
      <HttpStatusCode code={404} />
      <PageHeader
        title="Page not found"
        description="The page you're looking for doesn't exist or has moved."
      />
      <A href="/">Back to your items</A>
    </>
  );
}
