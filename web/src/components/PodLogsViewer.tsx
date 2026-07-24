import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import OptionSelect from "./OptionSelect";
import Notice from "./Notice";
import LastUpdateTime from "./LastUpdateTime";
import QueryStatus from "./QueryStatus";

interface LogViewerProps {
  namespace: string;
  podName: string;
  containers?: string[];
}

export default function PodLogsViewer({
  namespace,
  podName,
  containers,
}: LogViewerProps) {
  const [tailLines, setTailLines] = useState(100);
  const [container, setContainer] = useState(containers ? containers[0] : "");
  const [refetchIntervalSeconds, setRefetchIntervalSeconds] = useState(0);
  const url =
    `/api/v1/namespaces/${namespace}/pods/${podName}/logs` +
    `?tailLines=${tailLines}` +
    (container ? `&container=${container}` : "");
  const {
    data,
    dataUpdatedAt,
    isLoading,
    isLoadingError,
    isRefetchError,
    error,
  } = useQuery({
    queryKey: ["podLogs", podName, namespace, tailLines, container],
    queryFn: () => apiFetch(url, (r) => r.text()),
    refetchInterval: refetchIntervalSeconds * 1000,
  });

  return (
    <>
      <h2>Logs</h2>
      <LastUpdateTime timestamp={dataUpdatedAt} />
      <OptionSelect
        label="Container: "
        kind="string"
        value={container}
        changeHandler={setContainer}
        options={(containers && containers.map((c) => [c, c])) || []}
      ></OptionSelect>
      <OptionSelect
        label="Number of lines: "
        kind="number"
        value={tailLines}
        changeHandler={setTailLines}
        options={[
          ["10", 10],
          ["50", 50],
          ["100", 100],
          ["500", 500],
          ["1000", 1000],
        ]}
      />
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
      {data !== undefined &&
        (data.length > 0 ? (
          <pre>{data}</pre>
        ) : (
          <Notice type="info">
            Nothing to see here. The container may still be waiting to start.
          </Notice>
        ))}
    </>
  );
}
