type OtaLogoProps = {
  size?: number;
  className?: string;
  /** dark：侧栏/渐变底；light：白底卡片 */
  surface?: 'dark' | 'light';
};

/** OTA 侧栏标识：图标底板 + 设备 + 升级箭头 + 空口信号弧 */
export function OtaLogo({ size = 28, className, surface = 'dark' }: OtaLogoProps) {
  const onLight = surface === 'light';
  const plateFill = onLight ? 'rgba(37, 99, 169, 0.12)' : 'rgba(255,255,255,0.16)';
  const plateStroke = onLight ? 'rgba(37, 99, 169, 0.38)' : 'rgba(255,255,255,0.38)';
  const deviceFill = onLight ? 'rgba(37, 99, 169, 0.16)' : 'rgba(255,255,255,0.22)';

  return (
    <svg
      className={className}
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden
    >
      <rect
        x="4"
        y="4"
        width="24"
        height="24"
        rx="7"
        fill={plateFill}
        stroke={plateStroke}
        strokeWidth="1.25"
      />
      <path
        d="M7.5 13.5c4-4 13-4 17 0"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        opacity="0.45"
      />
      <path
        d="M10 15.5c3-2.5 9-2.5 12 0"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        opacity="0.72"
      />
      <rect
        x="11.25"
        y="16"
        width="9.5"
        height="8"
        rx="1.75"
        fill={deviceFill}
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M16 22.5V17.5M13.75 19.75 16 17.5l2.25 2.25"
        stroke="currentColor"
        strokeWidth="1.65"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
