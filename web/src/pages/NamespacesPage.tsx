import { useQuery } from "@tanstack/react-query";
import { useState, useMemo } from "react";
import { Link } from "@tanstack/react-router";
import { apiFetch } from "../api/client";
import InfoMessage from "../components/InfoMessage";

interface Namespace {
  name: string;
  status: string;
  created_at: string;
}

export default function NamespacesPage() {
  const url = `/api/v1/namespaces`;
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["namespaces"],
    queryFn: () => apiFetch<Namespace[]>(url, (r) => r.json()),
  });
  const [searchTerm, setSearchTerm] = useState("");
  const filteredData = useMemo(() => {
    return data?.filter((namespace) => namespace.name.includes(searchTerm));
  }, [searchTerm, data]);

  if (isLoading) return <>Loading...</>;
  if (isError) return <>Error: {error.message}</>;
  return (
    <>
      <input
        type="text"
        placeholder="Type to search..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
      ></input>
      {filteredData && filteredData.length > 0 ? (
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
        <InfoMessage>No namespaces to show.</InfoMessage>
      )}
    </>
  );
}
