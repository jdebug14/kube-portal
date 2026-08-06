export interface Condition {
  type: string;
  status: string;
  reason: string;
  message: string;
  last_update_time: string;
  last_transition_time: string;
}

export default function Conditions({
  conditions,
}: {
  conditions: Condition[];
}) {
  return (
    <ul className="badge-list">
      {conditions.map((condition) => (
        <li key={condition.type} className="badge badge-grey">
          <details>
            <summary>
              {condition.type}: {condition.status}
            </summary>
            <div className="condition-details">
              <p>{condition.message}</p>
              <p>Last transition time: {condition.last_transition_time}</p>
            </div>
          </details>
        </li>
      ))}
    </ul>
  );
}
