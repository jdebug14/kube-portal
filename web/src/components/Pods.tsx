import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { apiFetch } from "../api/client";
import Notice from "../components/Notice";
import QueryStatus from "../components/QueryStatus";

interface Pod {
  name: string;
  namespace: string;
  phase: string;
  created_at: string;
}

export default function Pods({ namespace }: { namespace: string }) {
  const url = `/api/v1/namespaces/${namespace}/pods`;
  const { data, isLoading, isLoadingError, isRefetchError, error } = useQuery({
    queryKey: ["pods", namespace],
    queryFn: () => apiFetch<Pod[]>(url, (r) => r.json()),
  });

  const [searchTerm, setSearchTerm] = useState("");
  const filteredData = useMemo(
    () => data?.filter((pod) => pod.name.includes(searchTerm)),
    [searchTerm, data],
  );

  return (
    <>
      <h3>Pods</h3>

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
                    to="/namespaces/$ns/pods/$name"
                    params={{ ns: namespace, name: pod.name }}
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
        </>
      )}
    </>
  );
}
