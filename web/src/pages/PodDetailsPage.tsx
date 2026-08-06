import { getRouteApi, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import KeyValueList from "../components/KeyValueList";
import EventsFeed from "../components/EventsFeed";
import PodLogsViewer from "../components/PodLogsViewer";
import LastUpdateTime from "../components/LastUpdateTime";
import QueryStatus from "../components/QueryStatus";

const routeApi = getRouteApi("/namespaces/$ns/pods/$name");

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
  const { ns, name } = routeApi.useParams();
  const url = `/api/v1/namespaces/${ns}/pods/${name}`;
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["podDetails", ns, name],
    queryFn: () => apiFetch<PodDetails>(url, (r) => r.json()),
  });

  const annotationEntries = data ? Object.entries(data.annotations ?? {}) : [];
  const labelEntries = data ? Object.entries(data.labels ?? {}) : [];
  return (
    <>
      <Link to="/namespaces/$ns" params={{ ns }}>
        ← namespaces/{ns}
      </Link>
      <h2>Pod: {name}</h2>
      <LastUpdateTime timestamp={dataUpdatedAt} />

      <QueryStatus
        isLoading={isLoading}
        isLoadingError={isLoadingError}
        isRefetchError={isRefetchError}
        error={error}
      />

      {data && (
        <>
          <h3>Properties</h3>
          <table className="resource-properties">
            <tbody>
              <tr>
                <td>Status</td>
                <td>{data.phase}</td>
              </tr>
              <tr>
                <td>Namespace</td>
                <td>{data.namespace}</td>
              </tr>
              <tr>
                <td>Node</td>
                <td>{data.host_node}</td>
              </tr>
              <tr>
                <td>Created</td>
                <td>{new Date(data.created_at).toLocaleString()}</td>
              </tr>
              <tr hidden={annotationEntries.length < 1}>
                <td>Annotations</td>
                <td>
                  <KeyValueList entries={annotationEntries} />
                </td>
              </tr>
              <tr hidden={labelEntries.length < 1}>
                <td>Labels</td>
                <td>
                  <KeyValueList entries={labelEntries} />
                </td>
              </tr>
            </tbody>
          </table>

          <h3>Containers</h3>
          {data.containers.map((container) => (
            <div key={container.name}>
              <h4>{container.name}</h4>
              <table className="resource-properties">
                <tbody>
                  <tr>
                    <td>Image</td>
                    <td>{container.image}</td>
                  </tr>
                  <tr>
                    <td>Ready</td>
                    <td>{String(container.ready)}</td>
                  </tr>
                  <tr>
                    <td>Restarts</td>
                    <td>{container.restarts}</td>
                  </tr>
                  {container.last_exit_time && (
                    <tr>
                      <td>Last terminiation</td>
                      <td>
                        Reason: {container.last_exit_reason}
                        <br />
                        Finished at:{" "}
                        {new Date(container.last_exit_time).toLocaleString()}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          ))}

          <EventsFeed namespace={ns} involvedObjectName={name} />

          <PodLogsViewer
            namespace={ns}
            podName={name}
            containers={data.containers.map((c) => c.name)}
          />
        </>
      )}
    </>
  );
}
