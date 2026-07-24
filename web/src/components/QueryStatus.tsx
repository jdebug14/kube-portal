import LoadingIndicator from "./LoadingIndicator";
import Notice from "./Notice";

type QueryStatusProps = {
  isLoading: boolean;
  isLoadingError: boolean;
  isRefetchError: boolean;
  error: Error | null;
};

export default function QueryStatus({
  isLoading,
  isLoadingError,
  isRefetchError,
  error,
}: QueryStatusProps) {
  if (isLoading) return <LoadingIndicator />;
  if (isLoadingError)
    return <Notice type="error">Error: {error!.message}</Notice>;
  if (isRefetchError)
    return (
      <Notice type="warning">
        Refresh failed - showing last known data ({error!.message})
      </Notice>
    );
  return null;
}
