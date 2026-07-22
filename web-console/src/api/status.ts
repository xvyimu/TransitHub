import axios from 'axios'
import { http } from './http'
import type { ApiResponse, ProbeName, ProbeResult, StatusData } from '@/types/api'

export async function getStatus() {
  const res = await http.get<ApiResponse<StatusData>>('/api/status', {
    skipAuthRedirect: true,
  })
  return res.data
}

async function probeOne(name: ProbeName, path: string): Promise<ProbeResult> {
  try {
    const res = await axios.get(path, {
      baseURL: '',
      withCredentials: true,
      validateStatus: () => true,
      timeout: 8000,
    })
    const ok = res.status >= 200 && res.status < 300
    return {
      name,
      ok,
      status: res.status,
      body: res.data,
    }
  } catch (e) {
    return {
      name,
      ok: false,
      status: null,
      body: null,
      error: e instanceof Error ? e.message : String(e),
    }
  }
}

/** Aggregate process probes (+ optional SPA edge health). */
export async function fetchProbes(): Promise<ProbeResult[]> {
  const jobs: Array<Promise<ProbeResult>> = [
    probeOne('healthz', '/healthz'),
    probeOne('livez', '/livez'),
    probeOne('readyz', '/readyz'),
    probeOne('frontend-healthz', '/frontend-healthz'),
  ]
  return Promise.all(jobs)
}
