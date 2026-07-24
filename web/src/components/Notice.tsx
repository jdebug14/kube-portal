import type { ReactNode } from "react";

type NoticeProps = {
  type: "info" | "warning" | "error";
  children: ReactNode;
};

export default function Notice({ type, children }: NoticeProps) {
  const icon = { info: "ℹ️", warning: "⚠️", error: "❌" }[type];
  return (
    <p>
      <span style={{ color: "blue" }}>{icon} </span> {children}
    </p>
  );
}
