import { getRouteApi, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import KeyValueList from "../components/KeyValueList";
import EventsFeed from "../components/EventsFeed";
import PodLogsViewer from "../components/PodLogsViewer";
import LastUpdateTime from "../components/LastUpdateTime";
import QueryStatus from "../components/QueryStatus";

const routeApi = getRouteApi("/namespaces/$ns/pods/$pn");

interface Container {
  name: string;
  image: string;
  ready: boolean;
  restarts: number;
  last_exit_time?: string;
  last_exit_reason?: string;
}

interface PodDetails {
  name: string;
  namespace: string;
  phase: string;
  host_node: string;
  created_at: string;
  annotations?: Record<string, string>;
  labels?: Record<string, string>;
  containers: Container[];
}

export default function PodDetailsPage() {
  const { ns, pn } = routeApi.useParams();
  const url = `/api/v1/namespaces/${ns}/pods/${pn}`;
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["podDetails", ns, pn],
    queryFn: () => apiFetch<PodDetails>(url, (r) => r.json()),
  });

  const annotationEntries = data ? Object.entries(data.annotations ?? {}) : [];
  const labelEntries = data ? Object.entries(data.labels ?? {}) : [];
  return (
    <>
      <Link to="/namespaces/$ns" params={{ ns }}>
        ← {ns}/Pods
      </Link>
      <h2>{pn}</h2>
      <LastUpdateTime timestamp={dataUpdatedAt} />

      <QueryStatus
        isLoading={isLoading}
        isLoadingError={isLoadingError}
        isRefetchError={isRefetchError}
        error={error}
      />

      {data && (
        <>
          <p>
            <strong>Status:</strong> {data.phase}
          </p>
          <p>
            <strong>Host node:</strong> {data.host_node}
          </p>
          <p>
            <strong>Created at:</strong> {data.created_at}
          </p>
          <KeyValueList title="Annotations" entries={annotationEntries} />
          <KeyValueList title="Labels" entries={labelEntries} />
          <p>
            <strong>Containers:</strong>
          </p>
          <ul>
            {data.containers.map((container) => (
              <li key={container.name}>
                Name:{container.name}
                <br />
                Image: {container.image}
                <br />
                Ready: {String(container.ready)}
                <br />
                Restarts: {container.restarts}
                <br />
                {container.last_exit_time && (
                  <>Last Termination At: {container.last_exit_time}</>
                )}
                <br />
                {container.last_exit_reason && (
                  <>Last Termination Reason: {container.last_exit_reason}</>
                )}
              </li>
            ))}
          </ul>
          <EventsFeed namespace={ns} involvedObjectName={pn} />
          <PodLogsViewer
            namespace={ns}
            podName={pn}
            containers={data.containers.map((c) => c.name)}
          />
        </>
      )}
    </>
  );
}
