import type { PropsWithChildren } from "react";

export default function InfoMessage({ children }: PropsWithChildren) {
  return (
    <p>
      <span style={{ color: "blue" }}>&#8505; </span> {children}
    </p>
  );
}
