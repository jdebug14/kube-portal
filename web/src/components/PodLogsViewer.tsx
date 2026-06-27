import { useState } from 'react';
import { useQuery } from '@tanstack/react-query'
import { apiFetch } from '../api/client';

interface LogViewerProps {
    namespace: string
    podName: string
    containers?: string[]
}

function PodLogsViewer({ namespace, podName, containers }: LogViewerProps) {
    const [tailLines, setTailLines] = useState(100)
    const [container, setContainer] = useState(containers ? containers[0] : '')
    const url = `/api/v1/namespaces/${namespace}/pods/${podName}/logs`
        + `?tailLines=${tailLines}`
        +  `&container=${container}`
    const { data, isLoading, isError, error } = useQuery({
        queryKey: ['podLogs', podName, namespace, tailLines, container],
        queryFn: () => apiFetch(url, r => r.text()),
    })

    return (
        <div>
            {isLoading && <div>Loading...</div>}
            {isError && <div>Error: {error.message}</div>}
            <h2>Logs</h2>
            Container: <select value={container} onChange={e => setContainer(e.target.value)}>
                {containers?.map(c => (
                    <option key={c} value={c}>{c}</option>
                ))}
            </select>
            Number of lines: <select value={tailLines} onChange={e => setTailLines(Number(e.target.value))}>
                <option value={10}>10</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
                <option value={500}>500</option>
                <option value={1000}>1000</option>
            </select>
            <pre>
                { data }
            </pre>
        </div>
    )
}

export default PodLogsViewer