import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { apiFetch } from '../api/client'

interface Namespace {
    name: string
    status: string
    created_at: string
}

export default function NamespacesPage() {
    const url = `/api/v1/namespaces`
    const { data, isLoading, isError, error } = useQuery({
        queryKey: ['namespaces'],
        queryFn: () => apiFetch<Namespace[]>(url, r => r.json()),
    })
    
    if (isLoading) return <>Loading...</>
    if (isError) return <>Error: {error.message}</>
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
