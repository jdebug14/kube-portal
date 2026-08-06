import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import OptionSelect from "./OptionSelect";
import Notice from "./Notice";
import LastUpdateTime from "./LastUpdateTime";
import QueryStatus from "./QueryStatus";

interface EventInvolvedObject {
  kind: string;
  name: string;
  namespace: string;
}

interface Event {
  type: string;
  reason: string;
  message: string;
  count: number;
  first_time: string;
  last_time: string;
  involved_object: EventInvolvedObject;
}

export default function EventsFeed({
  namespace,
  involvedObjectName,
}: {
  namespace: string;
  involvedObjectName?: string;
}) {
  const url =
    `/api/v1/namespaces/${namespace}/events` +
    (involvedObjectName ? `?involvedObjectName=${involvedObjectName}` : "");
  const [refetchIntervalSeconds, setRefetchIntervalSeconds] = useState(0);

  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["events", namespace, involvedObjectName],
    queryFn: () => apiFetch<Event[]>(url, (r) => r.json()),
    refetchInterval: refetchIntervalSeconds * 1000,
  });
  return (
    <>
      <h3>Events</h3>
      <LastUpdateTime timestamp={dataUpdatedAt} />
      <OptionSelect
        label="Refetch interval: "
        kind="number"
        value={refetchIntervalSeconds}
        changeHandler={setRefetchIntervalSeconds}
        options={[
          ["Never", 0],
          ["10 sec", 10],
          ["1 min", 60],
          ["5 min", 300],
        ]}
      />

      <QueryStatus
        isLoading={isLoading}
        isLoadingError={isLoadingError}
        isRefetchError={isRefetchError}
        error={error}
      />

      {data && (
        <>
          {data.length > 0 ? (
            <table className="events">
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Name</th>
                  <th>Reason</th>
                  <th>Message</th>
                  <th>Count</th>
                  <th>Last seen</th>
                </tr>
              </thead>
              <tbody>
                {data?.map((event, i) => (
                  <tr
                    key={`${event.involved_object.name}-${event.reason}-${i}`}
                  >
                    <td>{event.type}</td>
                    <td>{event.involved_object.name}</td>
                    <td>{event.reason}</td>
                    <td>{event.message}</td>
                    <td>{event.count}</td>
                    <td>{event.last_time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : (
            <Notice type="info">
              Nothing to see here. Events have a limited retention time.
            </Notice>
          )}
        </>
      )}
    </>
  );
}
