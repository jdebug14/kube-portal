import { useQuery } from '@tanstack/react-query'

interface EventInvolvedObject {
  kind: string
  name: string
  namespace: string
}

interface Event {
    type: string
    reason: string
    message: string
    count: number
    first_time: string
    last_time: string
    involved_object: EventInvolvedObject
}

const fetchEventsFeed = async (ns: string, invObjName?: string): Promise<Event[]> => {
    const url = `/api/v1/namespaces/${ns}/events`
    + (invObjName ? `?involvedObjectName=${invObjName}` : '')
    const res = await fetch(url)
    if (!res.ok) throw new Error('Network error');
    return res.json()
}

function EventsFeed({ namespace, involvedObjectName }: { namespace: string, involvedObjectName?: string }) {
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['events', namespace, involvedObjectName],
        queryFn: () => fetchEventsFeed(namespace, involvedObjectName),
    })
    return (
        <div>
            {isLoading && <div>Loading...</div>}
            {isError && <div>Error: {error.message}</div>}
            <h2>Events</h2>
            <ul>
                {data?.map(event => (
                    <li key={`${event.involved_object.name}-${event.reason}-${event.first_time}`}>
                        type={event.type}, name={event.involved_object.name}, reason={event.reason}, message={event.message}, count={event.count}, lastseen={event.last_time}
                    </li>
                ))}
            </ul>
        </div>
    )
}

export default EventsFeed