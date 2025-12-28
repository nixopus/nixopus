import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'

import { theme, useOpenapi } from 'vitepress-openapi/client'
import 'vitepress-openapi/dist/style.css'
import './style.css'

import spec from '../../src/openapi.json' assert { type: 'json' }

import InstallGenerator from '../components/InstallGenerator.vue'

export default {
    extends: DefaultTheme,
    async enhanceApp(ctx) {
        ctx.app.component('InstallGenerator', InstallGenerator)
        
        const openapi = useOpenapi({ spec,
            config: {
                server: {
                    allowCustomServer: true
                }
            }
         })
        ctx.app.provide('openapi', openapi)
        if (theme.enhanceApp) {
            await theme.enhanceApp({ ...ctx, openapi })
        }
    }
} satisfies Theme
