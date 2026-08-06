import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { apiFetch } from "../api/client";
import Notice from "./Notice";
import QueryStatus from "./QueryStatus";

interface Deployment {
  name: string;
  namespace: string;
  desired_replicas: number;
  ready_replicas: number;
  available_replicas: number;
  created_at: string;
}

export default function Deployments({ namespace }: { namespace: string }) {
  const url = `/api/v1/namespaces/${namespace}/deployments`;
  const { data, isLoading, isLoadingError, isRefetchError, error } = useQuery({
    queryKey: ["deployments", namespace],
    queryFn: () => apiFetch<Deployment[]>(url, (r) => r.json()),
  });

  const [searchTerm, setSearchTerm] = useState("");
  const filteredData = useMemo(
    () => data?.filter((deployment) => deployment.name.includes(searchTerm)),
    [searchTerm, data],
  );

  return (
    <>
      <h3>Deployments</h3>

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
              {filteredData.map((deployment) => (
                <li key={deployment.name}>
                  <Link
                    to="/namespaces/$ns/deployments/$name"
                    params={{ ns: namespace, name: deployment.name }}
                  >
                    {deployment.name}
                  </Link>
                  [{deployment.desired_replicas} / {deployment.ready_replicas}
                  {" ready"}]
                </li>
              ))}
            </ul>
          ) : (
            <Notice type="info">Nothing to see here.</Notice>
          )}
        </>
      )}
    </>
  );
}
