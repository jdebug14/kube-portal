function KeyValueList({ title, entries }: { title: string, entries: [string, string][] }) {
  return (
    <div>
      {title}:
      {entries.length > 0 ? (
        <ul>
          {entries.map(([key, value]) => (
            <li key={key}>{key}: {value}</li>
          ))}
        </ul>) : (<div>None</div>)
      }
    </div>
  )
}

export default KeyValueList