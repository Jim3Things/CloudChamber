import React, {useState} from "react"
import {Visibility, VisibilityOff} from "@material-ui/icons"
import {IconButton, InputAdornment, TextField} from "@material-ui/core"

// PasswordTextField implements a form text input field with password masking
// and the ability to turn the masking on and off.
//
// The masking starts as on.
export function PasswordTextField(props: {
    value: string,
    label: string,
    onChange: (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => void
}) {
    const [visible, setVisible] = useState(false)
    const onVisibleChange = () => setVisible(!visible)

    // Determine the visibility on / visibility off icon to show.
    const visibility = (isVisible: boolean) => {
        if (isVisible) {
            return <VisibilityOff/>
        }

        return <Visibility/>
    }

    return <TextField
        margin="dense"
        label={props.label}
        variant="outlined"
        type={!visible ? "password" : "text"}
        InputProps={{
            endAdornment: (
                <InputAdornment position="end">
                    <IconButton onClick={onVisibleChange}>
                        {visibility(visible)}
                    </IconButton>
                </InputAdornment>
            )
        }}
        value={props.value}
        onChange={props.onChange}
    />

}
