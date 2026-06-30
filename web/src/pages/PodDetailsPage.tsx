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
    <div>
      <Link to='/namespaces/$ns' params={{ ns }}>← {ns}/Pods</Link>
      {isLoading && <div>Loading...</div>}
      {isError && <div>Error: {error.message}</div>}
      {data && (<div>
        <h2>{pn}</h2>
        <div>Status: {data.phase}</div>
        <div>Host node: {data.host_node}</div>
        <div>Created at: {data.created_at}</div>
        <KeyValueList title='Annotations' entries={annotationEntries}/>
        <KeyValueList title='Labels' entries={labelEntries}/>
        <div>
          Containers:
          <ul>
            {data.containers.map(container => (
              <li key={container.name}>
                <div>Name:{container.name}</div>
                <div>Image: {container.image}</div>
                <div>Ready: {String(container.ready)}</div>
                <div>Restarts: {container.restarts}</div>
                {container.last_exit_time && <div>Last Termination At: {container.last_exit_time}</div>}
                {container.last_exit_reason && <div>Last Termination Reason: {container.last_exit_reason}</div>}
              </li>
            ))}
          </ul>
        </div>
        <EventsFeed namespace={ns} involvedObjectName={pn} />
        <PodLogsViewer namespace={ns} podName={pn} containers={data.containers.map(c => c.name)} />
      </div>)}
    </div>
  )
}

export default PodDetailsPage