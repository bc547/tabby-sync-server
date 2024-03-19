import component2 from '../components/component2.js'

export default {
    // name: 'Page2',
    // components: {component2},

    setup() {
        const {ref} = Vue;

        const info = ref({
            Version: "",
            Repo: "",
            Sha: "",
            BuildTime: ""
        })

        fetch('/health')
            .then(response => response.json())
            .then(data => info.value = data)
            .catch(error => console.error('Error:', error));

        return {
            info
        };
    },

    template: `
      <div class="text-lg mt-4">Server info:</div>
      <div class="mx-4">
        <table>
          <tr>
            <td>Version</td>
            <td>: {{info.Version}}</td>
          </tr>
          <tr>
            <td>RepoUrl</td>
            <td>: {{info.RepoUrl}}</td>
          </tr>
          <tr>
            <td>ShaCommit</td>
            <td>: {{info.ShaCommit}}</td>
          </tr>
          <tr>
            <td>BuildTime</td>
            <td>: {{info.BuildTime}}</td>
          </tr>
        </table>
      </div>
    `,
};