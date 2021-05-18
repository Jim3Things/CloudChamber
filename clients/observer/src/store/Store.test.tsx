import {renderHook} from "@testing-library/react-hooks"
import {Provider} from 'react-redux'
import {store, useAppDispatch} from "./Store"
import {StepperMode, TimeContext} from "../proxies/StepperProxy"
import {stepperSlice} from "./Store"

function wrapper(props: {
    children: any
}) {
    return <Provider store={store}>
        {props.children}
    </Provider>
}

it("should initialize correctly", () => {
    const {result} = renderHook(() =>
        store.getState(),
        {wrapper: wrapper })

    expect(result.current.cur.now).toBe(0)
    expect(result.current.cur.rate).toBe(0)
    expect(result.current.cur.mode).toBe(StepperMode.Paused)
})

it("should handle a single update", () => {
    const {result} = renderHook(() => {
        const dispatch = useAppDispatch()

        const newTime: TimeContext = {
            mode: StepperMode.Paused,
            now: 1,
            rate: 0

        }

        dispatch(stepperSlice.actions.updatePolicy(newTime))

        return store.getState()

    },
        {wrapper: wrapper })

    expect(result.current.cur.now).toBe(1)
    expect(result.current.cur.rate).toBe(0)
    expect(result.current.cur.mode).toBe(StepperMode.Paused)

})
