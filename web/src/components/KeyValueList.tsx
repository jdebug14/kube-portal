export default function KeyValueList({
  entries,
  variant = "badges",
}: {
  entries: [string, string][];
  variant?: "badges" | "compact";
}) {
  if (entries.length < 1) return <></>;
  return (
    <>
      <ul className={variant === "badges" ? "badge-list" : "compact-list"}>
        {entries.map(([key, value]) => (
          <li
            key={key}
            className={
              variant === "badges" ? "badge badge-grey" : "compact-item"
            }
          >
            {key}={value}
          </li>
        ))}
      </ul>
    </>
  );
}
