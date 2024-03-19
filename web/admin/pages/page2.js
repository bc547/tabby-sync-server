import component1 from '../components/component1.js'
import component2 from '../components/component2.js'

export default {
    // name: 'Page1',
    components: {component1,component2},

    setup() {
        const title = 'Page One'
        return {title}
    },

    template: `<div>
    <component1></component1>
    <Divider />
    <component2></component2>
</div>
    `,
};
