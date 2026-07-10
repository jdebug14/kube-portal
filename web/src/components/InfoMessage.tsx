export default function InfoMessage({ message }: { message: string }) {
  return (
    <p>
      <span style={{ color: "blue" }}>&#8505; </span> {message}
    </p>
  );
}
