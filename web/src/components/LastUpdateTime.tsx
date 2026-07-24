export default function LastUpdateTime({ timestamp }: { timestamp: number }) {
  if (timestamp <= 0) return null;
  return (
    <p style={{ fontStyle: "italic" }}>
      Last updated {new Date(timestamp).toLocaleTimeString()}
    </p>
  );
}
