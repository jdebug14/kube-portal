import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import OptionSelect from "./OptionSelect";
import InfoMessage from "./InfoMessage";

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

  const { data, isLoading, isFetching, isError, error } = useQuery({
    queryKey: ["events", namespace, involvedObjectName],
    queryFn: () => apiFetch<Event[]>(url, (r) => r.json()),
    refetchInterval: refetchIntervalSeconds * 1000,
  });
  return (
    <>
      <h2>Events</h2>
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
      {isLoading && <>Loading...</>}
      {isFetching && !isLoading && <>Refreshing...</>}
      {isError && <>Error: {error.message}</>}
      {data && data.length > 0 ? (
        <ul>
          {data?.map((event) => (
            <li
              key={`${event.involved_object.name}-${event.reason}-${event.first_time}`}
            >
              type={event.type}, name={event.involved_object.name}, reason=
              {event.reason}, message={event.message}, count={event.count},
              lastseen={event.last_time}
            </li>
          ))}
        </ul>
      ) : (
        <InfoMessage>
          No events to show. Events have a limited retention time.
        </InfoMessage>
      )}
    </>
  );
}
