# ILAYE (I'll Locate All Your Errors)

The trendiest new way to test your boards

## Demo

https://github.com/user-attachments/assets/d58d6cc9-528f-4ff0-93ff-97f0169b5764

## Commands

| Category         | Command Name             | Hex Code | Response Description                | Return Structure/Data | Notes                        |
| :--------------- | :----------------------- | :------- | :---------------------------------- | :-------------------- | :--------------------------- |
| **Connectivity** | LoRa Echo Test           | `0x02`   | `0x02` (Echo)                       | -                     | -                            |
| **Connectivity** | CAN Connectivity Test    | -        | -                                   | -                     | Performed via `0x00` request |
| **Mode**         | Normal Mode Transition   | `0x00`   | `0x00` (Ack)                        | -                     | -                            |
| **Mode**         | Inspect Mode Transition  | `0x01`   | `0x01` (Ack)                        | -                     | -                            |
| **Analog**       | Read Last Data Line      | `0xA0`   | File size + last recorded timestamp | `sd_card_update`      | -                            |
| **Analog**       | Sample LC Reading        | `0xA1`   | Sample 1 reading from LC            | `LC reading`          | Not used for Launch          |
| **Analog**       | Sample PT Readings       | `0xA2`   | Sample 1 set of readings from PTs   | `PT reading`          | Not used for TR              |
| **Analog**       | Sample INA Readings      | `0xA3`   | Sample 1 set of readings from INA   | `INA reading`         | -                            |
| **Digital**      | Read Last Data Line      | `0xB0`   | File size + last recorded timestamp | `sd_card_update`      | -                            |
| **Digital**      | Sample Altimeter Reading | `0xB1`   | Sample 1 reading from Altimeter     | `Altimeter readings`  | -                            |
| **Digital**      | Sample GPS Reading       | `0xB2`   | Sample 1 reading from GPS           | `GPS readings`        | -                            |
| **Digital**      | Sample Shock 1 Reading   | `0xB3`   | Sample 1 reading from Shock 1       | `Shock readings`      | -                            |
| **Digital**      | Sample Shock 2 Reading   | `0xB4`   | Sample 1 reading from Shock 2       | `Shock readings`      | -                            |
| **Digital**      | Sample IMU Reading       | `0xB5`   | Sample 1 reading from IMU           | `IMU Readings`        | -                            |
| **Resets**       | Reset Radio Board        | `0x0F`   | `0x0F` (Ack)                        | -                     | **Tentative**                |
| **Resets**       | Reset Analog Board       | `0xAF`   | `0xAF` (Ack)                        | -                     | **Tentative**                |
| **Resets**       | Reset Digital Board      | `0xBF`   | `0xBF` (Ack)                        | -                     | **Tentative**                |
| **Maintenance**  | Clear Analog SD Card     | `0xAE`   | `0xAE` (Ack)                        | -                     | -                            |
| **Maintenance**  | Clear Digital SD Card    | `0xBE`   | `0xBE` (Ack)                        | -                     | -                            |
