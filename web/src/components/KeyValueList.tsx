export default function KeyValueList({
  title,
  entries,
}: {
  title: string;
  entries: [string, string][];
}) {
  if (entries.length < 1) return null;
  return (
    <>
      <strong>{title}:</strong>
      <ul>
        {entries.map(([key, value]) => (
          <li key={key}>
            {key}: {value}
          </li>
        ))}
      </ul>
    </>
  );
}
