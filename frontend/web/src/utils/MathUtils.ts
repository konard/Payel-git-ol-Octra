/**
 * CrewAI - Math Utilities
 * Математические утилиты для работы с координатами и геометрией
 */

import type { Point, Rectangle, Size } from '../models/Types';

/**
 * Вычисляет расстояние между двумя точками
 * @param p1 Первая точка
 * @param p2 Вторая точка
 * @returns Расстояние между точками
 */
export function distance(p1: Point, p2: Point): number {
    const dx = p2.x - p1.x;
    const dy = p2.y - p1.y;
    return Math.sqrt(dx * dx + dy * dy);
}

/**
 * Проверяет, находится ли точка внутри прямоугольника
 * @param point Точка для проверки
 * @param rect Прямоугольник
 * @returns True если точка внутри прямоугольника
 */
export function isPointInRectangle(point: Point, rect: Rectangle): boolean {
    return (
        point.x >= rect.x &&
        point.x <= rect.x + rect.width &&
        point.y >= rect.y &&
        point.y <= rect.y + rect.height
    );
}

/**
 * Проверяет, находится ли точка внутри круга
 * @param point Точка для проверки
 * @param center Центр круга
 * @param radius Радиус круга
 * @returns True если точка внутри круга
 */
export function isPointInCircle(point: Point, center: Point, radius: number): boolean {
    return distance(point, center) <= radius;
}

/**
 * Находит ближайшую точку на линии к заданной точке
 * @param lineStart Начало линии
 * @param lineEnd Конец линии
 * @param point Точка для поиска ближайшей позиции
 * @returns Ближайшая точка на линии
 */
export function closestPointOnLine(lineStart: Point, lineEnd: Point, point: Point): Point {
    const dx = lineEnd.x - lineStart.x;
    const dy = lineEnd.y - lineStart.y;
    const lengthSquared = dx * dx + dy * dy;

    if (lengthSquared === 0) {
        return { ...lineStart };
    }

    let t = ((point.x - lineStart.x) * dx + (point.y - lineStart.y) * dy) / lengthSquared;
    t = Math.max(0, Math.min(1, t));

    return {
        x: lineStart.x + t * dx,
        y: lineStart.y + t * dy,
    };
}

/**
 * Вычисляет середину между двумя точками
 * @param p1 Первая точка
 * @param p2 Вторая точка
 * @returns Средняя точка
 */
export function midpoint(p1: Point, p2: Point): Point {
    return {
        x: (p1.x + p2.x) / 2,
        y: (p1.y + p2.y) / 2,
    };
}

/**
 * Вычисляет угол между двумя точками в радианах
 * @param p1 Первая точка
 * @param p2 Вторая точка
 * @returns Угол в радианах
 */
export function angleBetween(p1: Point, p2: Point): number {
    return Math.atan2(p2.y - p1.y, p2.x - p1.x);
}

/**
 * Поворачивает точку вокруг центра на заданный угол
 * @param point Точка для поворота
 * @param center Центр поворота
 * @param angle Угол поворота в радианах
 * @returns Повернутая точка
 */
export function rotatePoint(point: Point, center: Point, angle: number): Point {
    const cos = Math.cos(angle);
    const sin = Math.sin(angle);
    const dx = point.x - center.x;
    const dy = point.y - center.y;

    return {
        x: center.x + dx * cos - dy * sin,
        y: center.y + dx * sin + dy * cos,
    };
}

/**
 * Нормализует угол до диапазона [0, 2π)
 * @param angle Угол в радианах
 * @returns Нормализованный угол
 */
export function normalizeAngle(angle: number): number {
    angle = angle % (2 * Math.PI);
    if (angle < 0) {
        angle += 2 * Math.PI;
    }
    return angle;
}

/**
 * Интерполирует между двумя значениями
 * @param start Начальное значение
 * @param end Конечное значение
 * @param t Коэффициент интерполяции [0, 1]
 * @returns Интерполированное значение
 */
export function lerp(start: number, end: number, t: number): number {
    return start + (end - start) * t;
}

/**
 * Интерполирует между двумя точками
 * @param p1 Первая точка
 * @param p2 Вторая точка
 * @param t Коэффициент интерполяции [0, 1]
 * @returns Интерполированная точка
 */
export function lerpPoint(p1: Point, p2: Point, t: number): Point {
    return {
        x: lerp(p1.x, p2.x, t),
        y: lerp(p1.y, p2.y, t),
    };
}

/**
 * Ограничивает значение в заданном диапазоне
 * @param value Значение для ограничения
 * @param min Минимальное значение
 * @param max Максимальное значение
 * @returns Ограниченное значение
 */
export function clamp(value: number, min: number, max: number): number {
    return Math.min(Math.max(value, min), max);
}

/**
 * Преобразует значение из одного диапазона в другой
 * @param value Значение для преобразования
 * @param inMin Минимум входного диапазона
 * @param inMax Максимум входного диапазона
 * @param outMin Минимум выходного диапазона
 * @param outMax Максимум выходного диапазона
 * @returns Преобразованное значение
 */
export function mapRange(
    value: number,
    inMin: number,
    inMax: number,
    outMin: number,
    outMax: number
): number {
    return ((value - inMin) * (outMax - outMin)) / (inMax - inMin) + outMin;
}

/**
 * Вычисляет размер прямоугольника, содержащего два прямоугольника
 * @param rect1 Первый прямоугольник
 * @param rect2 Второй прямоугольник
 * @returns Объединенный прямоугольник
 */
export function unionRectangles(rect1: Rectangle, rect2: Rectangle): Rectangle {
    const x = Math.min(rect1.x, rect2.x);
    const y = Math.min(rect1.y, rect2.y);
    const width = Math.max(rect1.x + rect1.width, rect2.x + rect2.width) - x;
    const height = Math.max(rect1.y + rect1.height, rect2.y + rect2.height) - y;

    return { x, y, width, height };
}

/**
 * Вычисляет пересечение двух прямоугольников
 * @param rect1 Первый прямоугольник
 * @param rect2 Второй прямоугольник
 * @returns Пересечение или null если нет пересечения
 */
export function intersectRectangles(rect1: Rectangle, rect2: Rectangle): Rectangle | null {
    const x = Math.max(rect1.x, rect2.x);
    const y = Math.max(rect1.y, rect2.y);
    const width = Math.min(rect1.x + rect1.width, rect2.x + rect2.width) - x;
    const height = Math.min(rect1.y + rect1.height, rect2.y + rect2.height) - y;

    if (width <= 0 || height <= 0) {
        return null;
    }

    return { x, y, width, height };
}

/**
 * Проверяет пересечение двух прямоугольников
 * @param rect1 Первый прямоугольник
 * @param rect2 Второй прямоугольник
 * @returns True если прямоугольники пересекаются
 */
export function rectanglesIntersect(rect1: Rectangle, rect2: Rectangle): boolean {
    return intersectRectangles(rect1, rect2) !== null;
}

/**
 * Вычисляет точку на кривой Безье
 * @param p0 Начальная точка
 * @param p1 Первая контрольная точка
 * @param p2 Вторая контрольная точка
 * @param p3 Конечная точка
 * @param t Параметр [0, 1]
 * @returns Точка на кривой
 */
export function bezierPoint(p0: Point, p1: Point, p2: Point, p3: Point, t: number): Point {
    const oneMinusT = 1 - t;
    const oneMinusTSquared = oneMinusT * oneMinusT;
    const oneMinusTCubed = oneMinusTSquared * oneMinusT;
    const tSquared = t * t;
    const tCubed = tSquared * t;

    return {
        x:
            oneMinusTCubed * p0.x +
            3 * oneMinusTSquared * t * p1.x +
            3 * oneMinusT * tSquared * p2.x +
            tCubed * p3.x,
        y:
            oneMinusTCubed * p0.y +
            3 * oneMinusTSquared * t * p1.y +
            3 * oneMinusT * tSquared * p2.y +
            tCubed * p3.y,
    };
}

/**
 * Вычисляет длину кривой Безье методом дискретизации
 * @param p0 Начальная точка
 * @param p1 Первая контрольная точка
 * @param p2 Вторая контрольная точка
 * @param p3 Конечная точка
 * @param segments Количество сегментов для дискретизации
 * @returns Приближенная длина кривой
 */
export function bezierLength(
    p0: Point,
    p1: Point,
    p2: Point,
    p3: Point,
    segments: number = 20
): number {
    let length = 0;
    let previousPoint = p0;

    for (let i = 1; i <= segments; i++) {
        const t = i / segments;
        const point = bezierPoint(p0, p1, p2, p3, t);
        length += distance(previousPoint, point);
        previousPoint = point;
    }

    return length;
}
