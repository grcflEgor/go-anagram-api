import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const errorRate = new Rate('errors');

export const options = {
    stages: [
      { duration: '30s', target: 15 },   
      { duration: '30s', target: 20 },    
      { duration: '1m', target: 50 },    
      { duration: '30s', target: 0 },     
    ],
    thresholds: {
      http_req_duration: ['p(95)<1000'],  
      http_req_failed: ['rate<0.01'],     
      errors: ['rate<0.01'],
    },
  };
  

export default function () {
  const uploadData = {
    file: http.file('кот\nток\nрост\nторс\n', 'words.txt'),
    case_sensitive: 'false',
  };

  let uploadRes = http.post('http://localhost:8080/api/v1/anagrams/upload', uploadData);
  
  let ok = check(uploadRes, {
    'upload status is 202': (r) => r.status === 202,
    'upload response time < 2000ms': (r) => r.timings.duration < 2000,
  })
  if (!ok) {
    errorRate.add(1);
  }

  if (uploadRes.status === 202) {
    try {
      const response = JSON.parse(uploadRes.body);
      if (response.task_id) {
        sleep(0.2);
        
        let resultRes = http.get(`http://localhost:8080/api/v1/anagrams/groups/${response.task_id}`);
        
        ok = check(resultRes, {
          'get result status is 200 or 202': (r) => r.status === 200 || r.status === 202,
          'get result response time < 1000ms': (r) => r.timings.duration < 1000,
        })
        if (!ok) {
          errorRate.add(1);
        }
      }
    } catch (e) {
    }
  }

  const groupData = JSON.stringify({
    words: ['кот', 'ток', 'рост', 'торс'],
    case_sensitive: false
  });

  const groupParams = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  let groupRes = http.post('http://localhost:8080/api/v1/anagrams/group', groupData, groupParams);
  
  ok = check(groupRes, {
    'group status is 202': (r) => r.status === 202,
    'group response time < 1000ms': (r) => r.timings.duration < 1000,
  })
  if (!ok) {
    errorRate.add(1);
  }

  sleep(0.2);
}
