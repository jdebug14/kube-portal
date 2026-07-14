import InfoMessage from "./InfoMessage";

export default function KeyValueList({
  title,
  entries,
}: {
  title: string;
  entries: [string, string][];
}) {
  return (
    <>
      <strong>{title}:</strong>
      {entries.length > 0 ? (
        <ul>
          {entries.map(([key, value]) => (
            <li key={key}>
              {key}: {value}
            </li>
          ))}
        </ul>
      ) : (
        <InfoMessage>None</InfoMessage>
      )}
    </>
  );
}
