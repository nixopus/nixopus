"use client"
import ApplicationLogs from '@/app/self-host/components/application-details/logs'
import { useGetApplicationDeploymentByIdQuery } from '@/redux/services/deploy/applicationsApi'
import { useParams } from 'next/navigation'
import React from 'react'

function page() {
  const { deployment_id } = useParams()
  const { data: deployment } = useGetApplicationDeploymentByIdQuery({ id: deployment_id?.toString() as string }, { skip: !deployment_id })

  return (
    <div className='container mx-auto py-6 space-y-8 max-w-4xl 2xl:max-w-7xl'>
      <div className='mt-5 flex-col flex mb-4'>
        <span className='text-2xl font-bold'>Deployment Details</span>
        {/* <span className=''>status : {deployment?.status?.status}</span> */}
      </div>
      <ApplicationLogs logs={deployment?.logs} onRefresh={() => { }} currentPage={1} setCurrentPage={(page: number) => { }} />
    </div>
  )
}

export default page