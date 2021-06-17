// UsePopup is a simple helper hook that provides the common handling for
// triggering a popup display.

import React, {MouseEventHandler, useState} from "react"

export function usePopup<S>() : [boolean, S | null, MouseEventHandler<S>, MouseEventHandler<S>] {
    const [anchorEl, setAnchorEl] = useState<S | null>(null)

    const handlePopoverOpen = (event: React.MouseEvent<S, MouseEvent>): void => {
        setAnchorEl(event.currentTarget)
    }

    const handlePopoverClose = () => {
        setAnchorEl(null)
    }

    return [Boolean(anchorEl), anchorEl, handlePopoverOpen, handlePopoverClose]
}
