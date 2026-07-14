import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import OptionSelect from "./OptionSelect";
import InfoMessage from "./InfoMessage";

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
    `&container=${container}`;
  const { data, isLoading, isFetching, isError, error } = useQuery({
    queryKey: ["podLogs", podName, namespace, tailLines, container],
    queryFn: () => apiFetch(url, (r) => r.text()),
    refetchInterval: refetchIntervalSeconds * 1000,
  });

  return (
    <>
      <h2>Logs</h2>
      {isLoading && <>Loading...</>}
      {isFetching && !isLoading && <>Refreshing...</>}
      {isError && <>Error: {error.message}</>}
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
      ></OptionSelect>
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
      ></OptionSelect>
      {data ? (
        <pre>{data}</pre>
      ) : (
        <InfoMessage>
          No logs to show. The container may still be waiting to start.
        </InfoMessage>
      )}
    </>
  );
}
