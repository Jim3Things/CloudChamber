import React, {Component} from 'react';

interface Props {

}

interface State {

}

export class FaultInjectionPanel extends Component<Props, State> {
    componentWillMount() {

    }


    render() {
        return (
            <div>
                Fault Injection Functions
                <br/>
                <table>
                    <tbody>
                        <tr>
                            <td>list of active injections goes here</td>
                            <td>operations: new..., edit..., stop..., resume..., delete...</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        );
    }
}
