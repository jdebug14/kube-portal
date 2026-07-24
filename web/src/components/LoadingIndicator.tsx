export default function LoadingIndicator() {
  return (
    <>
      <span
        role="status"
        aria-label="Loading content"
        style={{
          display: "inline-block",
          width: "1em",
          height: "1em",
          border: "2px solid currentColor",
          borderTopColor: "transparent",
          borderRadius: "50%",
          animation: "spin 0.6s linear infinite",
        }}
      />
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </>
  );
}
