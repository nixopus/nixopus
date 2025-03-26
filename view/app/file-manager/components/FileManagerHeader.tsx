import React from 'react';

function FileManagerHeader() {
  return (
    <div>
      <div className="">
        <h1 className="text-md font-bold capitalize leading-normal text-primary sm:text-lg md:text-xl lg:text-3xl">
          File Manager
        </h1>
        <h2 className="text-xs leading-normal text-muted-foreground sm:text-sm lg:text-xl">
          Manage your files here
        </h2>
      </div>
    </div>
  );
}

export default FileManagerHeader;
