/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    container: {
      center: true,
      padding: '2rem',
      screens: {
        '2xl': '1400px',
      },
    },
    extend: {
      fontFamily: {
        sans: ['var(--font-inter)', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
        display: ['var(--font-inter)', 'system-ui', 'sans-serif'],
        japanese: ['Noto Sans JP', 'var(--font-inter)', 'sans-serif'],
      },
      colors: {
        // Hexabase Design System Colors
        // Brand Colors
        black: '#000000',
        white: '#FFFFFF',

        // Hexa Green Scale (Primary Accent)
        'hexa-green': {
          DEFAULT: '#00C6AB',
        },

        // Hexa Pink Scale (Secondary Accent)
        'hexa-pink': '#FF346B',

        // System Colors
        gray: {
          DEFAULT: '#CCCCCC',
          cancel: '#9e9e9e',
          'cancel-hover': '#b5b5b5',
          disabled: '#e6e6e6',
        },

        // Text Colors
        text: {
          primary: '#FFFFFF',
          placeholder: '#E6E6E6',
        },

        // Background Colors
        background: {
          DEFAULT: '#28292D', // Dark default
          sidebar: '#333336', // Thin/Side Bar
        },

        // Border Colors
        'hexa-border': {
          DEFAULT: '#555558',
          hover: '#656569',
          disabled: '#3D3D3F',
        },

        // Input Colors
        'hexa-input': {
          DEFAULT: '#38383B',
          hover: '#3A3A3E',
          disabled: '#2B2B2E',
        },

        // Error/Delete Colors
        error: {
          DEFAULT: '#FF7979',
          hover: '#FF9B9B',
        },

        // Legacy color mappings for compatibility
        primary: {
          DEFAULT: '#00C6AB',
          50: '#D4FFF9',
          100: '#AAFFF3',
          200: '#09FFDD',
          300: '#00F0CF',
          400: '#00DABC',
          500: '#00C6AB',
          600: '#00B29A',
          700: '#00A08B',
          800: '#00907D',
          900: '#00907D',
        },
        danger: {
          DEFAULT: '#FF7979',
          50: '#FFF0F0',
          100: '#FFE0E0',
          200: '#FFC0C0',
          300: '#FFA0A0',
          400: '#FF9B9B',
          500: '#FF7979',
          600: '#FF5555',
          700: '#FF3333',
          800: '#FF1111',
          900: '#DD0000',
        },
        // Keep original color variables for compatibility
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: "hsl(var(--foreground))",
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
      },
      // Hexabase Spacing System (8px base)
      spacing: {
        'xs': '2px',
        'sm': '4px',
        'md': '8px',
        'st': '16px',
        'lg': '24px',
        'xl': '28px',
        'xxl': '32px',
        'xxxl': '40px',
        'plus': '48px',
        'extended': '56px',
        'super': '64px',
        'queen': '72px',
        'king': '80px',
      },
      // Hexabase Elevation System
      boxShadow: {
        'dp-0': 'none',
        'dp-1': '0px 1px 2px 0px rgba(0, 0, 0, 0.16)',
        'dp-2': '0px 2px 3px 0px rgba(0, 0, 0, 0.16)',
        'dp-4': '0px 4px 5px 0px rgba(0, 0, 0, 0.16)',
        'dp-8': '0px 8px 9px 0px rgba(0, 0, 0, 0.15)',
        'dp-16': '0px 16px 17px 0px rgba(0, 0, 0, 0.15)',
        'dp-24': '0px 24px 25px 0px rgba(0, 0, 0, 0.04)',
        // Keep some legacy shadows for compatibility
        'sm': '0px 1px 2px 0px rgba(0, 0, 0, 0.16)',
        'DEFAULT': '0px 2px 3px 0px rgba(0, 0, 0, 0.16)',
        'md': '0px 4px 5px 0px rgba(0, 0, 0, 0.16)',
        'lg': '0px 8px 9px 0px rgba(0, 0, 0, 0.15)',
        'xl': '0px 16px 17px 0px rgba(0, 0, 0, 0.15)',
        '2xl': '0px 24px 25px 0px rgba(0, 0, 0, 0.04)',
      },
      // Typography
      fontSize: {
        // Headings
        'heading-xxl': ['24px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-xl': ['21px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-lg': ['18px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-base': ['16px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-md': ['14px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-sm': ['12px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        'heading-xs': ['10px', { lineHeight: '1.0', letterSpacing: '0.05em', fontWeight: '700' }],
        // Body Text
        'body-base': ['16px', { lineHeight: '2.0', letterSpacing: '0.05em', fontWeight: '400' }],
        'body-md': ['14px', { lineHeight: '2.0', letterSpacing: '0.03em', fontWeight: '400' }],
        'body-sm': ['12px', { lineHeight: '1.5', letterSpacing: '0.03em', fontWeight: '400' }],
        'body-xs': ['10px', { lineHeight: '1.5', letterSpacing: '0.03em', fontWeight: '400' }],
        // Keep default sizes for compatibility
        'xs': ['0.75rem', { lineHeight: '1rem' }],
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],
        'base': ['1rem', { lineHeight: '1.5rem' }],
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],
        '2xl': ['1.5rem', { lineHeight: '2rem' }],
      },
      lineHeight: {
        'label': '1.0',
        'description': '1.5',
        'note': '1.8',
        'body': '2.0',
      },
      letterSpacing: {
        'heading': '0.05em',
        'body': '0.03em',
        'body-wide': '0.05em',
      },
      borderRadius: {
        'sm': '0.25rem',
        'DEFAULT': '0.375rem',
        'md': '0.5rem',
        'lg': '0.75rem',
        'xl': '1rem',
        '2xl': '1.5rem',
        lg: `var(--radius)`,
        md: `calc(var(--radius) - 2px)`,
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}