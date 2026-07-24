import { useQuery } from "@tanstack/react-query";
import { useState, useMemo } from "react";
import { Link } from "@tanstack/react-router";
import { apiFetch } from "../api/client";
import LastUpdateTime from "../components/LastUpdateTime";
import Notice from "../components/Notice";
import QueryStatus from "../components/QueryStatus";

interface Namespace {
  name: string;
  status: string;
  created_at: string;
}

export default function NamespacesPage() {
  const url = `/api/v1/namespaces`;
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["namespaces"],
    queryFn: () => apiFetch<Namespace[]>(url, (r) => r.json()),
  });

  const [searchTerm, setSearchTerm] = useState("");
  const filteredData = useMemo(
    () => data?.filter((namespace) => namespace.name.includes(searchTerm)),
    [searchTerm, data],
  );

  return (
    <>
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
              {filteredData.map((ns) => (
                <li key={ns.name}>
                  <Link to="/namespaces/$ns" params={{ ns: ns.name }}>
                    {ns.name}
                  </Link>
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
