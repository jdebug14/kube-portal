import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import KeyValueList from '../components/KeyValueList'
import EventsFeed from '../components/EventsFeed'
import PodLogsViewer from '../components/PodLogsViewer'
import { apiFetch } from '../api/client'

const routeApi = getRouteApi('/namespaces/$ns/pods/$pn')

interface Container {
  name: string
  image: string
  ready: boolean
  restarts: number
  last_exit_time?: string
  last_exit_reason?: string
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

function PodDetailsPage() {
  const { ns, pn } = routeApi.useParams()
  const url = `/api/v1/namespaces/${ns}/pods/${pn}`
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['podDetails', ns, pn],
        queryFn: () => apiFetch<PodDetails>(url, r => r.json()),
    })

  const annotationEntries = data ? Object.entries(data.annotations ?? {}) : []
  const labelEntries = data ? Object.entries(data.labels ?? {}) : []

  return (
    <>
      <Link to='/namespaces/$ns' params={{ ns }}>← {ns}/Pods</Link>
      {isLoading && <p>Loading...</p>}
      {isError && <p>Error: {error.message}</p>}
      {data && (
        <>
        <h2>{pn}</h2>
        <p>Status: {data.phase}</p>
        <p>Host node: {data.host_node}</p>
        <p>Created at: {data.created_at}</p>
        <KeyValueList title='Annotations' entries={annotationEntries}/>
        <KeyValueList title='Labels' entries={labelEntries}/>
        <p>Containers:</p>
          <ul>
            {data.containers.map(container => (
              <li key={container.name}>
                Name:{container.name}
                <br/>Image: {container.image}
                <br/>Ready: {String(container.ready)}
                <br/>Restarts: {container.restarts}
                <br/>{container.last_exit_time && <>Last Termination At: {container.last_exit_time}</>}
                <br/>{container.last_exit_reason && <>Last Termination Reason: {container.last_exit_reason}</>}
              </li>
            ))}
          </ul>
        <EventsFeed namespace={ns} involvedObjectName={pn} />
        <PodLogsViewer namespace={ns} podName={pn} containers={data.containers.map(c => c.name)} />
        </>
      )}
    </>
  )
}

export default PodDetailsPage