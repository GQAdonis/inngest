import React from 'react';

import type { IconProps } from './props';

export default ({ size = '1em', fill = 'currentColor' }: IconProps) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <path
      d="M4 19.001H2C2 15.6873 4.68629 13.001 8 13.001C11.3137 13.001 14 15.6873 14 19.001H12C12 16.7918 10.2091 15.001 8 15.001C5.79086 15.001 4 16.7918 4 19.001ZM18.364 15.364L16.95 13.95C18.2629 12.6372 19.0005 10.8566 19.0005 8.99999C19.0005 7.14335 18.2629 5.36276 16.95 4.04999L18.364 2.63599C21.8781 6.1506 21.8781 11.8484 18.364 15.363V15.364ZM15.535 12.536L14.121 11.12C15.2908 9.94872 15.2908 8.05126 14.121 6.87999L15.535 5.46299C17.4876 7.4156 17.4876 10.5814 15.535 12.534V12.536ZM8 12C5.79086 12 4 10.2091 4 7.99999C4 5.79085 5.79086 3.99999 8 3.99999C10.2091 3.99999 12 5.79085 12 7.99999C12 9.06085 11.5786 10.0783 10.8284 10.8284C10.0783 11.5786 9.06087 12 8 12ZM8 5.99999C6.9074 6.00109 6.01789 6.87883 6.00223 7.97133C5.98658 9.06382 6.85057 9.96669 7.94269 9.9991C9.03481 10.0315 9.95083 9.18148 10 8.08999V8.48999V7.99999C10 6.89542 9.10457 5.99999 8 5.99999Z"
      fill={fill}
    ></path>
  </svg>
);