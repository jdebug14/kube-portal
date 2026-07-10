type OptionSelectProps =
  | {
      label: string;
      kind: "string";
      value: string;
      changeHandler: (e: string) => void;
      options: [string, string][];
    }
  | {
      label: string;
      kind: "number";
      value: number;
      changeHandler: (e: number) => void;
      options: [string, number][];
    };

export default function OptionSelect({
  label,
  kind,
  value,
  changeHandler,
  options,
}: OptionSelectProps) {
  return (
    <>
      <label>
        {label}
        <select
          value={value}
          onChange={(e) => {
            const raw = e.target.value;
            // discriminated union pattern
            if (kind === "number") changeHandler(Number(raw));
            else changeHandler(raw);
          }}
        >
          {options.map(([key, value]) => (
            <option key={key} value={value}>
              {key}
            </option>
          ))}
        </select>
      </label>
    </>
  );
}
