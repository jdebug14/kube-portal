import { getRouteApi, Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { apiFetch } from "../api/client";
import EventsFeed from "../components/EventsFeed";
import Notice from "../components/Notice";
import LastUpdateTime from "../components/LastUpdateTime";
import QueryStatus from "../components/QueryStatus";

const routeApi = getRouteApi("/namespaces/$ns");

interface Pod {
  name: string;
  namespace: string;
  phase: string;
  host_node: string;
  created_at: string;
}

export default function WorkloadsPage() {
  const { ns } = routeApi.useParams();
  const url = `/api/v1/namespaces/${ns}/pods`;
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["pods", ns],
    queryFn: () => apiFetch<Pod[]>(url, (r) => r.json()),
  });

  const [searchTerm, setSearchTerm] = useState("");
  const filteredData = useMemo(
    () => data?.filter((pod) => pod.name.includes(searchTerm)),
    [searchTerm, data],
  );

  return (
    <>
      <Link to="/">← Namespaces</Link>
      <h2>{ns}</h2>
      <LastUpdateTime timestamp={dataUpdatedAt} />

      <QueryStatus
        isLoading={isLoading}
        isLoadingError={isLoadingError}
        isRefetchError={isRefetchError}
        error={error}
      />

      {filteredData && (
        <>
          <input
            type="text"
            placeholder="Type to search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          ></input>
          {filteredData.length > 0 ? (
            <ul>
              {filteredData.map((pod) => (
                <li key={pod.name}>
                  <Link
                    to="/namespaces/$ns/pods/$pn"
                    params={{ ns: ns, pn: pod.name }}
                  >
                    {pod.name}
                  </Link>{" "}
                  [{pod.phase}]
                </li>
              ))}
            </ul>
          ) : (
            <Notice type="info">Nothing to see here.</Notice>
          )}
          <EventsFeed namespace={ns} />
        </>
      )}
    </>
  );
}
