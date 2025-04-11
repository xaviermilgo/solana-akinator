// src/components/Jinn.jsx
import React from 'react';
import jinnIdle from "../assets/jinn/state-idle.png";
import jinnThinking from "../assets/jinn/state-thinking.png";
import jinnAsking from "../assets/jinn/state-asking.png";
import jinnConfident from  "../assets/jinn/state-confident.png";
import jinnCorrect from  "../assets/jinn/state-correct.png";
import jinnWrong from  "../assets/jinn/state-wrong.png";
import jinnGlitched from  "../assets/jinn/state-glitched.png";

function Jinn({ state = 'idle' }) {
    const jinnImages = {
        'idle': jinnIdle,
        'thinking': jinnThinking,
        'asking': jinnAsking,
        'confident': jinnConfident,
        'correct': jinnCorrect,
        'wrong': jinnWrong,
        'glitched': jinnGlitched
    }
    const jinnState = Object.keys(jinnImages).includes(state) ? state : 'idle';
    const jinnImage = jinnImages[jinnState];

    return (
        <div className={`jinn jinn-${jinnState}`}>
            <img
                src={jinnImage}
                alt={`Jinn in ${jinnState} state`}
                className="jinn-image"
            />
        </div>
    );
}

export default Jinn;