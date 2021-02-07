import React from 'react';
import {Container, Item} from "../common/Cells";

export function WorkloadsPanel() {
    return (
        <Container>
            <Item xs={9}>
                Workload Functions
            </Item>
            <Item xs={9}>
                List of workloads goes here<br/>operations: new... update... stop... resume... delete...
            </Item>
        </Container>
    );
}
