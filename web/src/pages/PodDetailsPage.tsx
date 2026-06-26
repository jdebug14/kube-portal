import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import KeyValueList from '../components/KeyValueList'
import EventsFeed from '../components/EventsFeed'
import PodLogsViewer from '../components/PodLogsViewer'

const routeApi = getRouteApi('/namespaces/$ns/pods/$pn')

interface Container {
  name: string
  image: string
  ready: boolean
  restarts: number
  last_exit_code: number
  last_exit_reason: string
  last_finished_at: string
}

interface PodDetails {
    name: string
    namespace: string
    phase: string
    host_node: string
    created_at: string
    annotations?: Record<string, string>
    labels?: Record<string, string>
    containers: Container[]
}

const fetchPodDetails = async (ns: string, pn: string): Promise<PodDetails> => {
    const res = await fetch('/api/v1/namespaces/' + ns + '/pods/' + pn)
    if (!res.ok) throw new Error('Network error');
    return res.json()
}

function PodDetailsPage() {
  const { ns, pn } = routeApi.useParams()
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['podDetails', ns, pn],
        queryFn: () => fetchPodDetails(ns, pn),
    })

  const annotationEntries = data ? Object.entries(data.annotations ?? {}) : []
  const labelEntries = data ? Object.entries(data.labels ?? {}) : []

  return (
    <div>
      <Link to='/namespaces/$ns' params={{ ns }}>← {ns}/Pods</Link>
      {isLoading && <div>Loading...</div>}
      {isError && <div>Error: {error.message}</div>}
      <h2>{pn}</h2>
        <div>
            <div>Status: {data?.phase}</div>
            <div>Created at: {data?.created_at}</div>
            <KeyValueList title='Annotations' entries={annotationEntries}/>
            <KeyValueList title='Labels' entries={labelEntries}/>
            <div>
              Containers:
              <ul>
                {data?.containers.map(container => (
                  <li key={container.name}>
                    <div>Name:{container.name}</div>
                    <div>Image: {container.image}</div>
                    <div>Ready: {String(container.ready)}</div>
                    <div>Restarts: {container.restarts}</div>
                    <div>Last Exit Code: {container.last_exit_code}</div>
                    <div>Last Exit Reason: {container.last_exit_reason}</div>
                    <div>Last Finished At: {container.last_finished_at}</div>
                  </li>
                ))}
              </ul>
            </div>
            <EventsFeed namespace={ns} involvedObjectName={pn} />
            {data && (
              <PodLogsViewer namespace={ns} podName={pn} containers={data.containers.map(c => c.name)} />
            )}
          </div>
    </div>
  )
}

export default PodDetailsPage