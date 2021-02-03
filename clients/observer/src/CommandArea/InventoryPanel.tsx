import React, {Component} from 'react';

interface Props {

}
interface State {

}
export class InventoryPanel extends Component<Props, State> {
    componentWillMount() {

    }

    render() {
        return (
            <div>
                Inventory operations go here.<br/>
                Placeholder for future managed inventory change operations:
                    planned removal, planned maintenance, etc...
            </div>
        );
    }
}
