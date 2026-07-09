import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import EventsFeed from '../components/EventsFeed'
import { apiFetch } from '../api/client'

const routeApi = getRouteApi('/namespaces/$ns')

interface Pod {
    name: string
    namespace: string
    phase: string
    host_node: string
    created_at: string
}

export default function WorkloadsPage() {
  const { ns } = routeApi.useParams()
  const url = `/api/v1/namespaces/${ns}/pods`
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['pods', ns],
        queryFn: () => apiFetch<Pod[]>(url, r => r.json()),
    })
  const [ searchTerm, setSearchTerm ] = useState("");
  const filteredData = useMemo(() => { return data?.filter(pod => (pod.name.includes(searchTerm))) }, [searchTerm, data])

  return (
    <>
      <Link to="/">← Namespaces</Link>
      {isLoading && <>Loading...</>}
      {isError && <>Error: {error.message}</>}
      {data && (
        <>
        <h2>{ns}</h2>
        <input
          type='text'
          placeholder='Type to search...'
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}>
        </input>
        <ul>
          {filteredData?.map(pod => (
            <li key={pod.name}>
              <Link to="/namespaces/$ns/pods/$pn" params={{ns: ns, pn: pod.name}}>{pod.name}</Link> [{pod.phase}]
            </li>
          ))}
        </ul>
        <EventsFeed namespace={ns} />
        </>
      )}
    </>
  )
}
