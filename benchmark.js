const axios = require('axios').default

const configs = {
  'buran-remote': { baseUrl: 'http://content.example.com/api/v2' },
  'buran-local': { baseUrl: 'http://example-service/api/v2' },
  'prismic': { baseUrl: 'http://example-repo.cdn.prismic.io/api/v2' }
}

const { baseUrl, headers } = configs[process.argv[2]]

const query = ref => ({ params: { ref, q: '[[at(document.type, "sample_type")]]' } })

const measureOnce = () =>
  Promise.resolve({ start: Date.now() })
    .then(({ start }) =>
      axios.get(baseUrl, { headers })
        .then(({ data }) => data.refs[0].ref)
        .then(ref => axios.get(`${baseUrl}/documents/search`, { ...query(ref), headers }))
        .then(() => start))
    .then(start => Date.now() - start)

async function benchmark () {
  const times = []
  for (let i = 0; i < 10; i++) {
    times.push(await measureOnce())
    process.stdout.write('.')
  }

  console.log('')

  const result = {
    max: Math.max(...times),
    min: Math.min(...times),
    avg: times.reduce((sum, t) => sum + t, 0) / times.length
  }

  console.log(result)
}

benchmark()
