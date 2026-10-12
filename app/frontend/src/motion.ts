import { useReducedMotion } from "motion/react";
import type { MotionProps, Transition } from "motion/react";

/**
 * 动效令牌：全站共用同一套节奏，避免每个组件各写一个时长。
 * 控制类（按钮、卡片）要跟手；面板类（抽屉、卡片入场）要有重量；
 * 计量类（进度条、百分比）要慢，让人看清变化本身。
 */
export const springControl: Transition = { type: "spring", stiffness: 520, damping: 34, mass: 0.7 };
export const springPanel: Transition = { type: "spring", stiffness: 300, damping: 32, mass: 0.9 };
export const springMeter: Transition = { type: "spring", stiffness: 120, damping: 24 };
export const springPop: Transition = { type: "spring", stiffness: 640, damping: 26 };
export const fadeQuick: Transition = { duration: 0.18, ease: [0.22, 1, 0.36, 1] };

/** 列表逐项入场的间隔，超过 8 项后不再累加，避免最后一项等太久。 */
export const staggerStep = 0.045;

/**
 * 是否播放动效。系统"减少动效"或用户在设置里手动关闭时都返回 false。
 * 返回 false 时下方的预置不再给出初值，元素第一次渲染即为最终状态
 * ——这一点对测试很关键：断言可见性的元素不允许从 opacity:0 出发。
 */
export function useMotionEnabled() {
  return !useReducedMotion();
}

/** 入场：淡入并轻微上浮，用于卡片、段落、列表项。 */
export function rise(enabled: boolean, index = 0): MotionProps {
  if (!enabled) return {};
  return {
    initial: { opacity: 0, y: 16 },
    animate: { opacity: 1, y: 0 },
    transition: { ...springPanel, delay: Math.min(index, 8) * staggerStep },
  };
}

/** 纯淡入：用在不便位移的元素上，例如表格行、chips。 */
export function fadeIn(enabled: boolean, index = 0): MotionProps {
  if (!enabled) return {};
  return {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
    transition: { ...fadeQuick, delay: Math.min(index, 8) * staggerStep },
  };
}

/** 可点击卡片的悬停与按压反馈。位移交给 CSS 过渡，这里只做缩放，避免两套系统争抢 transform。 */
export function lift(enabled: boolean): MotionProps {
  if (!enabled) return {};
  return { whileHover: { scale: 1.006 }, whileTap: { scale: 0.994 }, transition: springControl };
}

/** 视图切换：key 变化时重新挂载，只播入场（不做退场等待，避免切换被拖慢）。 */
export function viewEnter(enabled: boolean): MotionProps {
  if (!enabled) return {};
  return {
    initial: { opacity: 0, y: 10 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.26, ease: [0.22, 1, 0.36, 1] },
  };
}

/** 抽屉与遮罩：需要退场动画，因此始终给出初始与退出值（减少动效时只留透明度）。 */
export const sheetVariants = {
  hidden: { opacity: 0, x: 40 },
  shown: { opacity: 1, x: 0, transition: springPanel },
  out: { opacity: 0, x: 24, transition: fadeQuick },
};

export const scrimVariants = {
  hidden: { opacity: 0 },
  shown: { opacity: 1, transition: fadeQuick },
  out: { opacity: 0, transition: fadeQuick },
};

/** 居中对话框：从下方轻微弹起。 */
export const dialogVariants = {
  hidden: { opacity: 0, y: 14, scale: 0.98 },
  shown: { opacity: 1, y: 0, scale: 1, transition: springPanel },
  out: { opacity: 0, y: 8, scale: 0.99, transition: fadeQuick },
};
