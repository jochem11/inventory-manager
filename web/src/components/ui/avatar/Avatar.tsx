import "./avatar.scss";

export type AvatarProps = {
  /** Full name; its initials are shown. */
  name: string;
  size?: "sm" | "md" | "lg";
  class?: string;
};

const initials = (name: string) =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]!.toUpperCase())
    .join("");

/** Round badge with someone's initials. Decorative — name them nearby. */
export function Avatar(props: AvatarProps) {
  return (
    <span class={`avatar avatar--${props.size ?? "md"} ${props.class ?? ""}`} aria-hidden="true">
      {initials(props.name)}
    </span>
  );
}
