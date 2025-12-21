import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'

import { theme, useOpenapi } from 'vitepress-openapi'
import 'vitepress-openapi/dist/style.css'
import './style.css'

import spec from '../../src/openapi.json' assert { type: 'json' }
import SponsorsMarquee from '../components/SponsorsMarquee.vue'

export default {
    extends: DefaultTheme,
    enhanceApp({ app, router, siteData }) {
        const openapi = useOpenapi({ spec })
        app.provide('openapi', openapi)

        // Register custom components
        app.component('SponsorsMarquee', SponsorsMarquee)

        theme.enhanceApp({ app })
    }
} satisfies Theme
