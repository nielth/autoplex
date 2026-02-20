import axios from "axios"
import { getApiDomain } from "../scripts/getApiDomain";
import { useEffect, useState } from "react";
import { formatBytes } from "../scripts/formatBytes";
import { authProvider } from "../auth";


export function DiskUsage() {
  const [disk, setDisk] = useState<DiskUsageObject>()
  const auth = authProvider

  useEffect(() => {
    const domain = getApiDomain();
    if (auth.isAuthenticated) {
      axios.get(`${domain}/api/disk`, { withCredentials: true }).then((resp) => {
        setDisk(resp.data.diskUsage)
      })
    }
  }, [auth.isAuthenticated]);

  return (<>
    {disk ?
      <div className="tooltip tooltip-bottom flex gap-4" data-tip="Storage left">
        <div className="w-max">
          <p className="font-bold">{formatBytes(disk['free'])}</p>
        </div>
        <progress
          className="progress progress-info hidden md:block w-56"
          value={disk['used']}
          max={disk['total']}
        />
      </div>
      : null
    }
  </>)
}
