import { getRouteApi, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import KeyValueList from "../components/KeyValueList";
import EventsFeed from "../components/EventsFeed";
import LastUpdateTime from "../components/LastUpdateTime";
import QueryStatus from "../components/QueryStatus";
import Conditions, { type Condition } from "../components/Conditions";

const routeApi = getRouteApi("/namespaces/$ns/deployments/$name");

interface DeploymentDetails {
  name: string;
  namespace: string;
  strategy: string;
  desired_replicas: number;
  updated_replicas: number;
  ready_replicas: number;
  available_replicas: number;
  rollout_status: string;
  conditions: Condition[];
  revisions: DeploymentRevision[];
  annotations?: Record<string, string>;
  labels?: Record<string, string>;
  created_at: string;
}

interface DeploymentRevision {
  name: string;
  number: number;
  is_current: boolean;
  desired_replicas: number;
  current_replicas: number;
  ready_replicas: number;
  pod_template: PodTemplate;
  created_at: string;
}

interface PodTemplate {
  containers: ContainerTemplate[];
  annotations?: Record<string, string>;
  labels?: Record<string, string>;
}

interface ContainerTemplate {
  name: string;
  image: string;
  command: string[];
  args: string[];
}

export default function DeploymentDetailsPage() {
  const { ns, name } = routeApi.useParams();
  const url = `/api/v1/namespaces/${ns}/deployments/${name}`;
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["deploymentDetails", ns, name],
    queryFn: () => apiFetch<DeploymentDetails>(url, (r) => r.json()),
  });

  const annotationEntries = data ? Object.entries(data.annotations ?? {}) : [];
  const labelEntries = data ? Object.entries(data.labels ?? {}) : [];
  return (
    <>
      <Link to="/namespaces/$ns" params={{ ns }}>
        ← namespaces/{ns}
      </Link>
      <h2>Deployment: {name}</h2>
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
                <td>{data.rollout_status}</td>
              </tr>
              <tr>
                <td>Strategy</td>
                <td>{data.strategy}</td>
              </tr>
              <tr>
                <td>Replicas</td>
                <td>
                  {data.desired_replicas} desired, {data.updated_replicas}{" "}
                  updated, {data.ready_replicas} ready,{" "}
                  {data.available_replicas} available
                </td>
              </tr>
              <tr>
                <td>Namespace</td>
                <td>
                  <Link to="/namespaces/$ns" params={{ ns: data.namespace }}>
                    {data.namespace}
                  </Link>
                </td>
              </tr>
              <tr>
                <td>Created</td>
                <td>{new Date(data.created_at).toLocaleString()}</td>
              </tr>
              <tr>
                <td>Annotations</td>
                <td>
                  <KeyValueList entries={annotationEntries} />
                </td>
              </tr>
              <tr>
                <td>Labels</td>
                <td>
                  <KeyValueList entries={labelEntries} />
                </td>
              </tr>
              <tr>
                <td>Conditions</td>
                <td>
                  <Conditions conditions={data.conditions} />
                </td>
              </tr>
            </tbody>
          </table>

          <h3>Revisions</h3>
          <table className="resource-properties">
            <thead>
              <tr>
                <th>#</th>
                <th>Pods</th>
                <th>Created</th>
                <th aria-label="Details" />
              </tr>
            </thead>
            <tbody>
              {data.revisions.map((revision) => (
                <tr key={revision.number}>
                  <td>{revision.number}</td>
                  <td>
                    {revision.desired_replicas}/{revision.ready_replicas}
                  </td>
                  <td>{new Date(revision.created_at).toLocaleString()}</td>
                  <td className="details-toggle">
                    <details>
                      <summary title="Show pod template details" />
                      <div className="pod-template-detail">
                        <p>Pod template:</p>
                        <div className="tab-2">
                          {revision.pod_template.annotations && (
                            <>
                              <p>Annotations:</p>
                              <KeyValueList
                                entries={Object.entries(
                                  revision.pod_template.annotations,
                                )}
                                variant="compact"
                              />
                            </>
                          )}
                          {revision.pod_template.labels && (
                            <>
                              <p>Labels:</p>
                              <KeyValueList
                                entries={Object.entries(
                                  revision.pod_template.labels,
                                )}
                                variant="compact"
                              />
                            </>
                          )}
                          <p>Containers:</p>
                          {revision.pod_template.containers.map((c) => (
                            <div key={c.name} className="tab-4">
                              {c.name}:
                              <div className="tab-6">
                                Image: {c.image}
                                {c.command && (
                                  <div>Command: {c.command.join(" ")}</div>
                                )}
                                {c.args && <div>Args: {c.args.join(" ")}</div>}
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    </details>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <EventsFeed namespace={ns} involvedObjectName={name} />
        </>
      )}
    </>
  );
}
