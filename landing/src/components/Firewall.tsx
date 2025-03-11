import React from 'react'
import { cn } from './lib/utils'
import { Activity, Clock, Network, Option, Plus } from 'lucide-react'

function Firewall() {
    return (
        <div className={cn("group text-white  p-8 md:p-12 lg:p-16")}>
            <div className='container mx-auto'>
                <h1 className='text-xl md:text-3xl font-bold flex items-center gap-4 font-mono mb-8'>
                    <Network className='w-12 h-12 animate-pulse text-yellow-400' />
                    <span className='bg-clip-text text-transparent bg-gradient-to-r from-yellow-400 to-red-400'>
                        Firewall
                    </span>
                </h1>
                <h2 className='text-xl md:text-xl font-semibold text-white mb-6 leading-tight'>
                    We know how important it is to secure our server/networks
                </h2>
                <ul className='space-y-6 text-lg'>
                    <li className='flex items-start'>
                        <Option className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Choose from variety of community built cron templates</span>
                    </li>
                    <li className='flex items-start'>
                        <Plus className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Create on your own</span>
                    </li>
                    <li className='flex items-start'>
                        <Activity className='w-6 h-6 text-yellow-400 mr-4 mt-1 flex-shrink-0' />
                        <span>Monitor and view cron logs effortlessly</span>
                    </li>
                </ul>
            </div>
        </div>
    )
}

export default Firewall