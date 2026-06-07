type OtaLogoProps = {
  size?: number;
  className?: string;
};

/** OTA 侧栏标识：设备块 + 向上箭头 + 广播弧（升级分发） */
export function OtaLogo({ size = 28, className }: OtaLogoProps) {
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
      <path
        d="M6 22c4.5-6.5 15.5-6.5 20 0"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        opacity="0.55"
      />
      <path
        d="M9 19c3-4 11-4 14 0"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        opacity="0.85"
      />
      <rect
        x="10.5"
        y="13"
        width="11"
        height="9.5"
        rx="2.2"
        stroke="currentColor"
        strokeWidth="1.75"
        fill="rgba(255,255,255,0.12)"
      />
      <path
        d="M16 20.5V14.5M13.2 16.8 16 14l2.8 2.8"
        stroke="currentColor"
        strokeWidth="1.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <circle cx="16" cy="11.2" r="1.2" fill="currentColor" />
    </svg>
  );
}
