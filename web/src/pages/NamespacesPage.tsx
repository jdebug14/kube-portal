import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { apiFetch } from '../api/client'

interface Namespace {
    name: string
    status: string
    created_at: string
}

function NamespacesPage() {
    const url = `/api/v1/namespaces`
    const { data, isLoading, isError, error } = useQuery({
        queryKey: ['namespaces'],
        queryFn: () => apiFetch<Namespace[]>(url, r => r.json()),
    })
    
    if (isLoading) return <div>Loading...</div>
    if (isError) return <div>Error: {error.message}</div>
    return (
        <ul>
            {data?.map(ns => (
                <li key={ns.name}>
                    <Link to="/namespaces/$ns" params={{ns: ns.name}} >{ns.name}</Link>
                </li>
            ))}
        </ul>
    )
}

export default NamespacesPage