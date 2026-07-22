import type { VariantProps } from 'class-variance-authority'
import { cva } from 'class-variance-authority'

export { default as Button } from './Button.vue'

export const buttonVariants = cva(
  'focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 rounded-full border border-transparent bg-clip-padding text-xs font-black tracking-tight uppercase focus-visible:ring-3 aria-invalid:ring-3 active:scale-[0.98] [&_svg:not([class*=size-])]:size-3.5 group/button inline-flex shrink-0 items-center justify-center whitespace-nowrap transition-all outline-none select-none disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-xs',
        outline: 'border-border/80 border-dashed bg-background hover:bg-muted hover:text-foreground dark:bg-input/30 dark:border-input dark:hover:bg-input/50 aria-expanded:bg-muted aria-expanded:text-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80 aria-expanded:bg-secondary aria-expanded:text-secondary-foreground',
        ghost: 'hover:bg-muted hover:text-foreground dark:hover:bg-muted/50 aria-expanded:bg-muted aria-expanded:text-foreground',
        destructive: 'bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/20 focus-visible:ring-rose-500/20 dark:bg-rose-500/20',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        'default': 'h-8 gap-1.5 px-3 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2',
        'xs': 'h-6 gap-1 px-2 text-[10px] font-black uppercase tracking-wider [&_svg:not([class*=size-])]:size-3',
        'sm': 'h-7 gap-1 px-2.5 text-xs font-black uppercase tracking-wider [&_svg:not([class*=size-])]:size-3.5',
        'lg': 'h-9 gap-1.5 px-4 text-xs font-black uppercase tracking-widest',
        'icon': 'size-8 rounded-full',
        'icon-xs': 'size-6 rounded-full [&_svg:not([class*=size-])]:size-3',
        'icon-sm': 'size-7 rounded-full',
        'icon-lg': 'size-9 rounded-full',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)
export type ButtonVariants = VariantProps<typeof buttonVariants>
