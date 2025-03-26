import React from 'react';
import { Activity, Clock, Option, Plus } from 'lucide-react';
import { cn } from './lib/utils';

function CronJob() {
    return (
        <div className={cn("group text-white  p-8 md:p-12 lg:p-16")}>

            <div className='container mx-auto'>
                <h1 className='text-xl md:text-3xl font-bold flex items-center gap-4 font-mono mb-8'>
                    <Clock className='w-12 h-12 animate-pulse text-yellow-400' />
                    <span className='bg-clip-text text-transparent bg-gradient-to-r from-yellow-400 to-red-400'>
                        Cron Jobs
                    </span>
                </h1>
                <h2 className='text-xl md:text-xl font-semibold text-white mb-6 leading-tight'>
                    We know that you&apos;re busy. It&apos;s super fast to run cron jobs.
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
    );
}

export default CronJob 